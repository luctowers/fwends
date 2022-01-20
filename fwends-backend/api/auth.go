package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"fwends-backend/config"
	"fwends-backend/handler"
	"io"
	"net/http"

	"github.com/go-redis/redis/v8"
	"google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

// GET /api/auth/config
//
// Get authentication configuration.
func AuthConfig(cfg *config.Config) handler.Handler {
	// contruct response
	var resbody struct {
		Enable   bool `json:"enable"`
		Services struct {
			GoogleClientID string `json:"google,omitempty"`
		} `json:"services"`
	}
	resbody.Enable = cfg.Auth.Enable
	resbody.Services.GoogleClientID = cfg.Auth.GoogleClientID

	// convert the response to bytes prior to request as it is static
	bytes, err := json.Marshal(resbody)
	if err != nil {
		panic(err)
	}

	return handler.NewStaticHandler(bytes, "application/json", http.StatusOK)
}

// GET /api/auth
//
// Checks whether the current session is authenticated.
func AuthVerify(cfg *config.Config, rdb *redis.Client) handler.Handler {
	if !cfg.Auth.Enable {
		return handler.NewErrorHandler(
			http.StatusMisdirectedRequest, errors.New("authentication is not enabled"),
		)
	} else {
		return &authVerifyHandler{cfg, rdb}
	}
}

type authVerifyHandler struct {
	cfg *config.Config
	rdb *redis.Client
}

func (h *authVerifyHandler) Handle(i handler.Input) (int, error) {
	// determine authentication status via redis
	var authenticated bool
	session, err := i.Request.Cookie(h.cfg.Auth.SessionCookie)
	if err != nil { // cookie not found
		authenticated = false
	} else {
		key := h.cfg.Auth.SessionsRedisPrefix + session.Value
		exists, err := h.rdb.Exists(i.Request.Context(), key).Result()
		if err != nil {
			return http.StatusInternalServerError, err
		} else {
			authenticated = exists == 1
		}
	}

	// respond with a boolean value
	i.Response.Header().Set("Content-Type", "application/json")
	if authenticated {
		i.Response.Write([]byte("true"))
	} else {
		i.Response.Write([]byte("false"))
	}

	return http.StatusOK, nil
}

// POST /api/auth
//
// Receives a token from the user, aunticates it and creates a session.
func Authenticate(cfg *config.Config, db *sql.DB, rdb *redis.Client) handler.Handler {
	if !cfg.Auth.Enable {
		return handler.NewErrorHandler(
			http.StatusMisdirectedRequest, errors.New("authentication is not enabled"),
		)
	} else {
		svc, err := newAuthServices(context.Background(), cfg)
		if err != nil {
			panic(err)
		}
		return &authenticateHandler{cfg, db, rdb, svc}
	}
}

type authenticateHandler struct {
	cfg *config.Config
	db  *sql.DB
	rdb *redis.Client
	svc *authServices
}

func (h *authenticateHandler) Handle(i handler.Input) (int, error) {
	// decode request body
	decoder := json.NewDecoder(i.Request.Body)
	var reqbody struct {
		Token   string `json:"token"`
		Service string `json:"service"`
	}
	err := decoder.Decode(&reqbody)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err)
	}

	// get verified email from token
	var email string
	switch reqbody.Service {
	case "google":
		email, err = getEmailFromGoogleToken(reqbody.Token, h.svc.google)
		if err != nil {
			return http.StatusBadRequest, err
		}
	default:
		return http.StatusBadRequest, fmt.Errorf("unrecognized auth service: %v", reqbody.Service)
	}

	// check whether email is admin
	rows, err := h.db.QueryContext(i.Request.Context(), "SELECT 1 FROM admins WHERE email = $1", email)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer rows.Close()
	if !rows.Next() {
		// no row was returned, so the email is not admin
		return http.StatusUnauthorized, errors.New("unauthorized authentication attempt")
	}

	// create session
	id := make([]byte, h.cfg.Auth.SessionIDSize)
	n, err := io.ReadFull(rand.Reader, id[:])
	if err != nil {
		return http.StatusInternalServerError, err
	} else if n != h.cfg.Auth.SessionIDSize {
		return http.StatusInternalServerError, errors.New("session id generation failed")
	}
	idB64 := base64.StdEncoding.EncodeToString(id[:])
	key := h.cfg.Auth.SessionsRedisPrefix + idB64
	val, err := h.rdb.SetNX(i.Request.Context(), key, true, h.cfg.Auth.SessionTTL).Result()
	if err != nil {
		return http.StatusInternalServerError, err
	} else if !val {
		return http.StatusInternalServerError, errors.New("session id collision")
	}

	// everything succeeded
	sessionCookie := http.Cookie{
		Name:     h.cfg.Auth.SessionCookie,
		Value:    idB64,
		MaxAge:   int(h.cfg.Auth.SessionTTL.Seconds()),
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(i.Response, &sessionCookie)

	return http.StatusOK, nil
}

// HELPERS

type authServices struct {
	google *oauth2.Service
}

func newAuthServices(ctx context.Context, cfg *config.Config) (*authServices, error) {
	services := &authServices{}
	if cfg.Auth.GoogleClientID != "" {
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
