package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fwends-backend/util"
	"net/http"
	"regexp"
	"sync"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

/*
Example curl commands:

curl -X POST http://localhost:8080/api/packs/ -d '{"title":"Test Pack"}'
curl -X GET http://localhost:8080/api/packs/6882582496895041536
curl -X PUT http://localhost:8080/api/packs/6882582496895041536 -d '{"title":"Updated Test Pack"}'
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536
curl -X PUT http://localhost:8080/api/packs/6882582496895041536/role/string -H 'Content-Type: image/png' --data-binary "@path/to/image.png"
curl -X PUT http://localhost:8080/api/packs/6882582496895041536/role/string -H 'Content-Type: audio/mpeg' --data-binary "@path/to/audio.mp3"
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536/role/string
*/

var identifierExpression = regexp.MustCompile(`^[a-z0-9_]{1,63}$`)

type packBody struct {
	Title string `json:"title"`
}

type idResponse struct {
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
			_, err := db.ExecContext(r.Context(), "INSERT INTO packs (id, title) VALUES ($1, $2)", id, body.Title)
			if err != nil {
				util.Error(w, http.StatusInternalServerError)
				log.WithError(err).Error("Failed to insert new pack")
			} else {
				w.Header().Set("Content-Type", "application/json")
				var response idResponse
				response.ID = id
				json.NewEncoder(w).Encode(response)
			}
		}
	}
}

func GetPack(db *sql.DB) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("pack_id")
		rows, err := db.QueryContext(r.Context(), "SELECT title FROM packs WHERE id = $1;", id)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to get pack")
		} else {
			defer rows.Close()
			if !rows.Next() {
				// no row was returned
				util.Error(w, http.StatusNotFound)
				log.Warn("Failed to get pack, it probably doesn't exist")
			} else {
				var body packBody
				err := rows.Scan(&body.Title)
				if err != nil {
					util.Error(w, http.StatusInternalServerError)
					log.WithError(err).Error("Failed to scan pack")
				} else {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(body)
				}
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
			id := ps.ByName("pack_id")
			res, err := db.ExecContext(r.Context(), "UPDATE packs SET title = $2 WHERE id = $1;", id, body.Title)
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
		id := ps.ByName("pack_id")
		res, err := db.ExecContext(r.Context(), "DELETE FROM packs WHERE id = $1;", id)
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

func resourceKey(packID string, roleID string, stringID string, resourceClass string) string {
	return fmt.Sprintf(
		"packs/%s/%s/%s/%s",
		packID,
		roleID,
		stringID,
		resourceClass,
	)
}

func deriveResourceClass(contentType string) (string, error) {
	// detemine resource type from extension
	switch contentType {
	case "image/webp":
		fallthrough
	case "image/bmp":
		fallthrough
	case "image/gif":
		fallthrough
	case "image/jpeg":
		fallthrough
	case "image/png":
		fallthrough
	case "image/svg+xml":
		return "image", nil
	case "audio/aac":
		fallthrough
	case "audio/mpeg":
		fallthrough
	case "audio/ogg":
		fallthrough
	case "audio/opus":
		fallthrough
	case "audio/wav":
		fallthrough
	case "audio/webm":
		fallthrough
	case "audio/flac":
		return "audio", nil
	default:
		return "", errors.New("unsupported pack resource content type")
	}
}

func UploadPackResource(db *sql.DB, s3c *s3.Client) httprouter.Handle {
	bucket := viper.GetString("s3_media_bucket")
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// get path params
		packID := ps.ByName("pack_id")
		roleID := ps.ByName("role_id")
		stringID := ps.ByName("string_id")

		// validation
		if !identifierExpression.MatchString(roleID) {
			util.Error(w, http.StatusBadRequest)
			log.WithFields(log.Fields{
				"roleId": roleID,
			}).Warn("Failed to validate role id")
			return
		}
		if !identifierExpression.MatchString(stringID) {
			util.Error(w, http.StatusBadRequest)
			log.WithFields(log.Fields{
				"stringId": stringID,
			}).Warn("Failed to validate string id")
			return
		}

		// determine whether upload is audio or image
		contentType := r.Header.Get("Content-Type")
		resourceClass, err := deriveResourceClass(contentType)
		if err != nil {
			util.Error(w, http.StatusBadRequest)
			log.WithError(err).WithFields(log.Fields{
				"contentType": contentType,
			}).Warn("Failed derive resource class")
			return
		}

		// begin a new transcation
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to begin postgres transcation")
			return
		}
		defer tx.Rollback()

		// check whether resource already exists, and wether is ready state
		rows, err := tx.QueryContext(
			r.Context(),
			"SELECT ready FROM packResources WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND class = $4",
			packID, roleID, stringID, resourceClass,
		)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to check if pack resource exists")
			return
		}
		resourceExists := rows.Next()
		var resourceReady bool
		if resourceExists {
			rows.Scan(&resourceReady)
		}
		rows.Close()

		if !resourceExists {

			// insert new resource
			_, err := tx.ExecContext(
				r.Context(),
				"INSERT INTO packResources (packId, roleId, stringID, class, ready) VALUES ($1, $2, $3, $4, FALSE)",
				packID, roleID, stringID, resourceClass,
			)
			if err != nil {
				util.Error(w, http.StatusInternalServerError)
				log.WithError(err).Error("Failed to insert new pack resource")
				return
			}

		} else if resourceReady {

			// update existing resource to be not ready
			result, err := tx.ExecContext(
				r.Context(),
				"UPDATE packResources SET ready = FALSE WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND class = $4",
				packID, roleID, stringID, resourceClass,
			)
			if err != nil {
				util.Error(w, http.StatusInternalServerError)
				log.WithError(err).Error("Failed to mark pack resource as not ready")
				return
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected != 1 {
				util.Error(w, http.StatusInternalServerError)
				log.WithFields(log.Fields{
					"rowsAffected": rowsAffected,
				}).Error("Failed to mark pack resource as not ready")
				return
			}

		}

		// commit current transaction and start a new one
		err = tx.Commit()
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to commit pack resource delete")
			return
		}
		tx, err = db.BeginTx(r.Context(), nil)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to begin postgres transaction")
			return
		}
		defer tx.Rollback()

		// indicate resource is ready (not commited yet), goal is to lock the row
		result, err := tx.ExecContext(
			r.Context(),
			"UPDATE packResources SET ready = TRUE WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND class = $4 AND ready = FALSE",
			packID, roleID, stringID, resourceClass,
		)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed update pack resource to ready state")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 1 {
			util.Error(w, http.StatusInternalServerError)
			log.WithFields(log.Fields{
				"rowsAffected": rowsAffected,
			}).Error("Failed update pack resource to ready state")
			return
		}

		// upload resource, now that the row is lock
		key := resourceKey(packID, roleID, stringID, resourceClass)
		_, err = s3c.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:        &bucket,
			Key:           &key,
			Body:          r.Body,
			ContentLength: r.ContentLength,
			ContentType:   &contentType,
		}, s3.WithAPIOptions(
			v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
		))
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to upload new pack resource")
			return
		} else {
			log.WithFields(log.Fields{
				"bucket": bucket,
				"key":    key,
			}).Info("Uploaded new pack resource")
		}

		// commit current transaction
		err = tx.Commit()
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to commit pack upload")
			return
		}

		util.OK(w)

	}
}

func DeletePackString(db *sql.DB, s3c *s3.Client) httprouter.Handle {
	bucket := viper.GetString("s3_media_bucket")
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// get path params
		packID := ps.ByName("pack_id")
		roleID := ps.ByName("role_id")
		stringID := ps.ByName("string_id")

		// begin a new transcation
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to begin postgres transcation")
			return
		}
		defer tx.Rollback()

		// update resources to be not ready state
		result, err := tx.ExecContext(
			r.Context(),
			"UPDATE packResources SET ready = FALSE WHERE packId = $1 AND roleId = $2 AND stringId = $3",
			packID, roleID, stringID,
		)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to mark pack resource as not ready")
			return
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			util.Error(w, http.StatusBadRequest)
			log.WithFields(log.Fields{
				"rowsAffected": rowsAffected,
			}).Error("Failed to mark pack resource as not ready")
			return
		}

		// commit current transaction and start a new one
		err = tx.Commit()
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to commit pack resource delete")
			return
		}
		tx, err = db.BeginTx(r.Context(), nil)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to begin postgres transaction")
			return
		}
		defer tx.Rollback()

		// delete resource rows (not commited yet)
		_, err = tx.ExecContext(
			r.Context(),
			"DELETE FROM packResources WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND ready = FALSE",
			packID, roleID, stringID,
		)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to mark pack resource as not ready")
			return
		}

		// delete objects from s3
		resourceClasses := []string{"image", "audio"}
		cerr := make(chan error, len(resourceClasses))
		var wg sync.WaitGroup
		for _, class := range resourceClasses {
			wg.Add(1)
			go func(class string) {
				defer wg.Done()
				key := resourceKey(packID, roleID, stringID, class)
				_, err := s3c.DeleteObject(r.Context(), &s3.DeleteObjectInput{
					Bucket: &bucket,
					Key:    &key,
				})
				if err != nil {
					cerr <- err
				}
			}(class)
		}
		wg.Wait()

		// handle errors produced by object deletes
	loop:
		for {
			select {
			case err := <-cerr:
				util.Error(w, http.StatusInternalServerError)
				log.WithError(err).Error("Failed to delete pack resource object")
				return
			default:
				break loop
			}
		}

		// commit current transaction
		err = tx.Commit()
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to commit pack upload")
			return
		}

		util.OK(w)

	}
}
