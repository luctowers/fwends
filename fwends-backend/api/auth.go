package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"fwends-backend/util"
	"io"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

// GET /api/auth/config
//
// Get authentication configuration.
func AuthConfig(logger *zap.Logger) httprouter.Handle {

	type authServiceInfo struct {
		GoogleClientId string `json:"google,omitempty"`
		// TODO: add more oauth services
	}

	type responseBody struct {
		Enable   bool            `json:"enable"`
		Services authServiceInfo `json:"services"`
	}

	// create auth info
	resbody := responseBody{
		Enable: viper.GetBool("auth_enable"),
		Services: authServiceInfo{
			GoogleClientId: viper.GetString("google_client_id"),
		},
	}

	// convert the response to bytes prior to request as it is static
	bytes, err := json.Marshal(resbody)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to encode auth info")
	}

	return util.WrapDecoratedHandle(
		logger,
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params, logger *zap.Logger) (int, error) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(bytes)
			return http.StatusOK, nil
		},
	)

}

// GET /api/auth
//
// Checks whether the current session is authenticated.
func AuthVerify(logger *zap.Logger, rdb *redis.Client) httprouter.Handle {

	if !viper.GetBool("auth_enable") {

		return util.WrapDecoratedHandle(
			logger,
			func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {
				return http.StatusMisdirectedRequest, errors.New("authentication is not enabled")
			},
		)

	} else {

		sessionCookie := viper.GetString("session_cookie")
		sessionRedisPrefix := viper.GetString("session_redis_prefix")

		return util.WrapDecoratedHandle(
			logger,
			func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {

				// determine authentication status via redis
				var authenticated bool
				session, err := r.Cookie(sessionCookie)
				if err != nil { // cookie not found
					authenticated = false
				} else {
					key := sessionRedisPrefix + session.Value
					exists, err := rdb.Exists(r.Context(), key).Result()
					if err != nil {
						return http.StatusInternalServerError, err
					} else {
						authenticated = exists == 1
					}
				}

				// respond
				w.Header().Set("Content-Type", "application/json")
				if authenticated {
					w.Write([]byte("true"))
				} else {
					w.Write([]byte("false"))
				}

				return http.StatusOK, nil

			},
		)

	}

}

// POST /api/auth
//
// Receives a token from the user, aunticates it and creates a session.
func Authenticate(logger *zap.Logger, db *sql.DB, rdb *redis.Client) httprouter.Handle {
	if !viper.GetBool("auth_enable") {

		return util.WrapDecoratedHandle(
			logger,
			func(w http.ResponseWriter, r *http.Request, _ httprouter.Params, logger *zap.Logger) (int, error) {
				return http.StatusMisdirectedRequest, errors.New("authentication is not enabled")
			},
		)

	} else {

		type requestBody struct {
			Token   string `json:"token"`
			Service string `json:"service"`
		}

		services, err := openAuthServices(context.Background())
		if err != nil {
			logger.With(zap.Error(err)).Fatal("Failed to open auth services")
		}

		sessionIDSize := viper.GetInt("session_id_size")
		sessionTTL := viper.GetDuration("session_ttl")
		sessionCookie := viper.GetString("session_cookie")
		sessionRedisPrefix := viper.GetString("session_redis_prefix")

		return util.WrapDecoratedHandle(
			logger,
			func(w http.ResponseWriter, r *http.Request, _ httprouter.Params, logger *zap.Logger) (int, error) {

				// decode request body
				decoder := json.NewDecoder(r.Body)
				var reqbody requestBody
				err := decoder.Decode(&reqbody)
				if err != nil {
					return http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err)
				}

				// get verified email from token
				var email string
				switch reqbody.Service {
				case "google":
					email, err = getEmailFromGoogleToken(reqbody.Token, services.google)
					if err != nil {
						return http.StatusBadRequest, err
					}
				default:
					return http.StatusBadRequest, fmt.Errorf("unrecognized auth service: %v", reqbody.Service)
				}

				// check whether email is admin
				rows, err := db.QueryContext(r.Context(), "SELECT 1 FROM admins WHERE email = $1", email)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				defer rows.Close()
				if !rows.Next() {
					// no row was returned, so the email is not admin
					return http.StatusUnauthorized, errors.New("unauthorized authentication attempt")
				}

				// create session
				id := make([]byte, sessionIDSize)
				n, err := io.ReadFull(rand.Reader, id[:])
				if err != nil {
					return http.StatusInternalServerError, err
				} else if n != sessionIDSize {
					return http.StatusInternalServerError, errors.New("session id generation failed")
				}
				idB64 := base64.StdEncoding.EncodeToString(id[:])
				key := sessionRedisPrefix + idB64
				val, err := rdb.SetNX(r.Context(), key, true, sessionTTL).Result()
				if err != nil {
					return http.StatusInternalServerError, err
				} else if !val {
					return http.StatusInternalServerError, errors.New("session id collision")
				}

				// everything succeeded
				sessionCookie := http.Cookie{
					Name:     sessionCookie,
					Value:    idB64,
					MaxAge:   int(sessionTTL.Seconds()),
					Secure:   true,
					HttpOnly: true,
				}
				http.SetCookie(w, &sessionCookie)

				return http.StatusOK, nil

			},
		)

	}
}

// HELPERS

type authServices struct {
	google *oauth2.Service
}

func openAuthServices(ctx context.Context) (*authServices, error) {
	services := &authServices{}
	if viper.IsSet("google_client_id") {
		google, err := oauth2.NewService(ctx, option.WithHTTPClient(&http.Client{}))
		if err != nil {
			return nil, err
		}
		services.google = google
	}
	return services, nil
}

func getEmailFromGoogleToken(token string, google *oauth2.Service) (string, error) {
	if google == nil {
		return "", errors.New("google oauth2 is not enabled")
	}
	call := google.Tokeninfo()
	call.IdToken(token)
	info, err := call.Do()
	if err != nil {
		return "", err
	}
	if !info.EmailVerified {
		return "", errors.New("google token info did not contain a verified email")
	}
	return info.Email, nil
}
