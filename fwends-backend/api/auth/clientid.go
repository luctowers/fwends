package auth

import (
	"encoding/json"
	"net/http"
	"os"

	"fwends-backend/util"

	"github.com/julienschmidt/httprouter"
)

func OauthClientId() httprouter.Handle {
	id := os.Getenv("OAUTH2_CLIENT_ID")
	if len(id) == 0 {
		return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			util.Error(w, http.StatusNoContent)
		}
	} else {
		return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			json.NewEncoder(w).Encode(id)
		}
	}
}
