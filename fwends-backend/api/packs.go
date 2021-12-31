package api

import (
	"database/sql"
	"encoding/json"
	"fwends-backend/util"
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

/*
Example curl commands:

curl -X POST http://localhost:8080/api/packs/ -d '{"title":"Test Pack"}'
curl -X PUT http://localhost:8080/api/packs/6882582496895041536 -d '{"title":"Updated Test Pack"}'
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536'
*/

type packBody struct {
	Title string `json:"title"`
}

type packResponse struct {
	ID int64 `json:"id,string"`
}

func CreatePack(db *sql.DB, snowflake *util.SnowflakeGenerator) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		decoder := json.NewDecoder(r.Body)
		var body packBody
		err := decoder.Decode(&body)
		if err != nil {
			util.Error(w, http.StatusBadRequest)
			log.WithError(err).Warn("Failed to decode pack body")
		} else if len(body.Title) == 0 {
			util.Error(w, http.StatusBadRequest)
			log.Warn("Empty pack title is not allowed")
		} else {
			id := snowflake.GenID()
			_, err := db.Exec("INSERT INTO packs (id, title) VALUES ($1, $2)", id, body.Title)
			if err != nil {
				util.Error(w, http.StatusInternalServerError)
				log.WithError(err).Error("Failed to insert new pack")
			} else {
				w.Header().Set("Content-Type", "application/json")
				var response packResponse
				response.ID = id
				json.NewEncoder(w).Encode(response)
			}
		}
	}
}

func UpdatePack(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		decoder := json.NewDecoder(r.Body)
		var body packBody
		err := decoder.Decode(&body)
		if err != nil {
			util.Error(w, http.StatusBadRequest)
			log.WithError(err).Warn("Failed to decode pack body")
		} else if len(body.Title) == 0 {
			util.Error(w, http.StatusBadRequest)
			log.Warn("Empty pack title is not allowed")
		} else {
			id := ps.ByName("id")
			res, err := db.Exec("UPDATE packs SET title = $2 WHERE id = $1;", id, body.Title)
			if err != nil {
				util.Error(w, http.StatusInternalServerError)
				log.WithError(err).Error("Failed to update pack")
			} else {
				affected, err := res.RowsAffected()
				if err != nil {
					util.Error(w, http.StatusInternalServerError)
					log.WithError(err).Error("Failed to get rows affected")
				} else if affected != 1 {
					util.Error(w, http.StatusBadRequest)
					log.WithFields(log.Fields{
						"rowsAffected": affected,
					}).Warn("Failed to update pack, it probably doesn't exist")
				} else {
					util.OK(w)
				}
			}
		}
	}
}

func DeletePack(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		res, err := db.Exec("DELETE FROM packs WHERE id = $1;", id)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to delete pack")
		} else {
			affected, err := res.RowsAffected()
			if err != nil {
				util.Error(w, http.StatusInternalServerError)
				log.WithError(err).Error("Failed to get rows affected")
			} else if affected != 1 {
				util.Error(w, http.StatusBadRequest)
				log.WithFields(log.Fields{
					"rowsAffected": affected,
				}).Warn("Failed to delete pack, it probably doesn't exist")
			} else {
				util.OK(w)
			}
		}
	}
}
