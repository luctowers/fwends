package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fwends-backend/util"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

// TODO: make these configurable
const sessionIDSize = 32          // 32 bytes
const sessionTTL = 24 * time.Hour // 1 day

type authServiceInfo struct {
	GoogleClientId string `json:"google,omitempty"`
	// TODO: add more oauth services
}

type authInfo struct {
	Enable   bool            `json:"enable"`
	Services authServiceInfo `json:"services"`
}

type authBody struct {
	Token   string `json:"token"`
	Service string `json:"service"`
}

type authServices struct {
	google *oauth2.Service
}

func AuthConfig() httprouter.Handle {
	info := authInfo{
		Enable: len(os.Getenv("AUTH_ENABLE")) != 0,
		Services: authServiceInfo{
			GoogleClientId: os.Getenv("GOOGLE_CLIENT_ID"),
		},
	}
	bytes, err := json.Marshal(info)
	if err != nil {
		log.WithError(err).Fatal("Failed to encode auth info")
	}
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	}
}

func Authenticate(db *sql.DB, rdb *redis.Client) httprouter.Handle {
	if len(os.Getenv("AUTH_ENABLE")) == 0 {
		return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			util.Error(w, http.StatusMisdirectedRequest)
		}
	} else {
		services := openAuthServices()
		return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			authenticateBody(w, r, db, rdb, services)
		}
	}
}

func openAuthServices() authServices {
	services := authServices{}
	var httpClient = &http.Client{}
	google, err := oauth2.NewService(context.Background(), option.WithHTTPClient(httpClient))
	if err != nil {
		log.WithError(err).Fatal("Failed to create Google oauth2 client")
	}
	services.google = google
	return services
}

func authenticateBody(w http.ResponseWriter, r *http.Request, db *sql.DB, rdb *redis.Client, services authServices) {
	decoder := json.NewDecoder(r.Body)
	var body authBody
	err := decoder.Decode(&body)
	if err != nil {
		util.Error(w, http.StatusBadRequest)
		log.WithError(err).Warn("Failed to decode authentication body")
	} else {
		switch body.Service {
		case "google":
			authenticateGoogleToken(w, body.Token, db, rdb, services.google)
		default:
			util.Error(w, http.StatusBadRequest)
			log.WithFields(log.Fields{
				"service": body.Service,
			}).Warn("Unrecognized auth service")
		}
	}
}

func authenticateGoogleToken(w http.ResponseWriter, token string, db *sql.DB, rdb *redis.Client, google *oauth2.Service) {
	if google == nil {
		util.Error(w, http.StatusBadRequest)
		log.Warn("Google oauth2 is not enabled")
	} else {
		call := google.Tokeninfo()
		call.IdToken(token)
		info, err := call.Do()
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Warn("Failed to get oauth2 id token info")
		} else {
			if !info.VerifiedEmail {
				util.Error(w, http.StatusInternalServerError)
				log.Warn("Oauth2 token info did not contain a verified email")
			} else {
				authenticateEmail(w, info.Email, db, rdb)
			}
		}
	}
}

func authenticateEmail(w http.ResponseWriter, email string, db *sql.DB, rdb *redis.Client) {
	rows, err := db.Query("SELECT 1 FROM admins WHERE email = $1", email)
	if err != nil {
		util.Error(w, http.StatusInternalServerError)
		log.WithError(err).Error("Failed to query postgres for admin email")
	} else if !rows.Next() {
		// no row was returned, so the email is not admin
		util.Error(w, http.StatusUnauthorized)
		log.Warn("Unauthorized authentication attempt")
	} else {
		authenticateCreateSession(w, rdb)
	}
}

func authenticateCreateSession(w http.ResponseWriter, rdb *redis.Client) {
	var id [sessionIDSize]byte
	n, err := io.ReadFull(rand.Reader, id[:])
	if err != nil || n != sessionIDSize {
		util.Error(w, http.StatusInternalServerError)
		log.WithError(err).Error("Failed to create secure session id")
	} else {
		id := base64.StdEncoding.EncodeToString(id[:])
		key := "session/" + id
		val, err := rdb.SetNX(context.Background(), key, true, sessionTTL).Result()
		if err != nil || !val {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to set session key in redis")
		} else {
			cookie := http.Cookie{
				Name:   "fwends_session",
				Value:  id,
				MaxAge: int(sessionTTL.Seconds()),
			}
			http.SetCookie(w, &cookie)
			util.Ok(w)
		}
	}
}
