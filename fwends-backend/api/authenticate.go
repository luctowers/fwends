package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"fwends-backend/api/util"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/oauth2/v2"
)

type authenticationBody struct {
	Token *string `json:"token"`
}

// TODO: make these configurable
const sessionIDSize = 32          // 32 bytes
const sessionTTL = 24 * time.Hour // 1 day

func Authenticate(db *sql.DB, rdb *redis.Client, oauth2 *oauth2.Service) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		decoder := json.NewDecoder(r.Body)
		var body authenticationBody
		err := decoder.Decode(&body)
		if err != nil {
			util.Error(w, http.StatusBadRequest)
			log.WithError(err).Warn("Failed to decode authentication body")
		} else if body.Token == nil {
			util.Error(w, http.StatusBadRequest)
			log.WithError(err).Warn("Missing token in authentication body")
		} else {
			authenticateToken(w, *body.Token, db, rdb, oauth2)
		}
	}
}

func authenticateToken(w http.ResponseWriter, token string, db *sql.DB, rdb *redis.Client, oauth2 *oauth2.Service) {
	call := oauth2.Tokeninfo()
	call.IdToken(token)
	tokenInfo, err := call.Do()
	if err != nil {
		util.Error(w, http.StatusInternalServerError)
		log.WithError(err).Warn("Failed to get oauth2 id token info")
	} else {
		authenticateTokenInfo(w, tokenInfo, db, rdb)
	}
}

func authenticateTokenInfo(w http.ResponseWriter, tokenInfo *oauth2.Tokeninfo, db *sql.DB, rdb *redis.Client) {
	if !tokenInfo.VerifiedEmail {
		util.Error(w, http.StatusInternalServerError)
		log.Warn("Oauth2 token info did not contain a verified email")
	} else {
		rows, err := db.Query("SELECT 1 FROM admins WHERE email = $1", tokenInfo.Email)
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
				Name:   "fwends-session",
				Value:  id,
				MaxAge: int(sessionTTL.Seconds()),
			}
			http.SetCookie(w, &cookie)
			util.Ok(w)
		}
	}
}
