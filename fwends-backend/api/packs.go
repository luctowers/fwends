package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fwends-backend/util"
	"net/http"
	"regexp"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

/*
Example curl commands:

curl -X POST http://localhost:8080/api/packs/ -d '{"title":"Test Pack"}'
curl -X GET http://localhost:8080/api/packs/6882582496895041536
curl -X PUT http://localhost:8080/api/packs/6882582496895041536 -d '{"title":"Updated Test Pack"}'
curl -X PUT http://localhost:8080/api/packs/6882582496895041536/role/string -H 'Content-Type: image/png' --data-binary "@path/to/image.png"
curl -X PUT http://localhost:8080/api/packs/6882582496895041536/role/string -H 'Content-Type: audio/mpeg' --data-binary "@path/to/audio.mp3"
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536/role
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
		defer rows.Close()
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
			if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == "23503" {
				// 23503 is foreign key constraint violation, meaning the pack doesn't exist
				util.Error(w, http.StatusNotFound)
				log.WithError(err).Warn("Failed to insert new pack resource, pack probably does not exist")
				return
			} else if err != nil {
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

func DeletePack(db *sql.DB, s3c *s3.Client) httprouter.Handle {
	bucket := viper.GetString("s3_media_bucket")
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// get path params
		packID := ps.ByName("pack_id")

		// delete pack resources
		err := deleteRelatedResources(r.Context(), db, s3c, &bucket, &packID, nil, nil)
		if err != nil {
			if err != nil {
				switch err.(type) {
				case *noResourcesFoundError:
					// a pack is allowed to have no resources
				default:
					util.Error(w, http.StatusInternalServerError)
					log.WithError(err).Error("Failed to delete resource in pack")
					return
				}
			}
		}

		// delete pack
		res, err := db.ExecContext(r.Context(), "DELETE FROM packs WHERE id = $1;", packID)
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to delete pack")
			return
		}

		// check how many rows were affected
		affected, err := res.RowsAffected()
		if err != nil {
			util.Error(w, http.StatusInternalServerError)
			log.WithError(err).Error("Failed to get rows affected")
			return
		} else if affected != 1 {
			util.Error(w, http.StatusNotFound)
			log.WithFields(log.Fields{
				"rowsAffected": affected,
			}).Warn("Failed to delete pack, it probably doesn't exist")
			return
		}

		util.OK(w)

	}
}

func DeletePackRole(db *sql.DB, s3c *s3.Client) httprouter.Handle {
	bucket := viper.GetString("s3_media_bucket")
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// get path params
		packID := ps.ByName("pack_id")
		roleID := ps.ByName("role_id")

		err := deleteRelatedResources(r.Context(), db, s3c, &bucket, &packID, &roleID, nil)
		if err != nil {
			if err != nil {
				switch err.(type) {
				case *noResourcesFoundError:
					util.Error(w, http.StatusNotFound)
					log.WithError(err).Warn("No resources found to delete in pack role")
					return
				default:
					util.Error(w, http.StatusInternalServerError)
					log.WithError(err).Error("Failed to delete resource in pack role")
					return
				}
			}
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

		err := deleteRelatedResources(r.Context(), db, s3c, &bucket, &packID, &roleID, &stringID)
		if err != nil {
			if err != nil {
				switch err.(type) {
				case *noResourcesFoundError:
					util.Error(w, http.StatusNotFound)
					log.WithError(err).Warn("No resources found to delete in pack string")
					return
				default:
					util.Error(w, http.StatusInternalServerError)
					log.WithError(err).Error("Failed to delete resource in pack string")
					return
				}
			}
		}

		util.OK(w)

	}
}

// HELPERS

type noResourcesFoundError struct{}

func (e *noResourcesFoundError) Error() string {
	return "no pack resources found"
}

func deleteRelatedResources(ctx context.Context, db *sql.DB, s3c *s3.Client, bucket *string, packID *string, roleID *string, stringID *string) error {

	const (
		deletePackMode   = 0
		deleteRoleMode   = 1
		deleteStringMode = 2
	)

	var mode int
	switch {
	case packID != nil && roleID == nil && stringID == nil:
		mode = deletePackMode
	case packID != nil && roleID != nil && stringID == nil:
		mode = deleteRoleMode
	case packID != nil && roleID != nil && stringID != nil:
		mode = deleteStringMode
	default:
		return errors.New("invalid arguments in call to deleteRelatedResources")
	}

	// begin a new transcation
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// update resources to be not ready state
	var rows *sql.Rows
	switch mode {
	case deletePackMode:
		rows, err = tx.QueryContext(
			ctx, "UPDATE packResources SET ready = FALSE WHERE packId = $1 RETURNING roleId, stringId, class",
			*packID,
		)
	case deleteRoleMode:
		rows, err = tx.QueryContext(
			ctx, "UPDATE packResources SET ready = FALSE WHERE packId = $1 AND roleId = $2 RETURNING roleId, stringId, class",
			*packID, *roleID,
		)
	case deleteStringMode:
		rows, err = tx.QueryContext(
			ctx, "UPDATE packResources SET ready = FALSE WHERE packId = $1 AND roleId = $2 AND stringId = $3 RETURNING roleId, stringId, class",
			*packID, *roleID, *stringID,
		)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	// determine which resources need to be deleted
	type resourceID struct {
		roleID   string
		stringID string
		class    string
	}
	resourceIDs := []resourceID{}
	for rows.Next() {
		var id resourceID
		err := rows.Scan(&id.roleID, &id.stringID, &id.class)
		if err != nil {
			return err
		}
		resourceIDs = append(resourceIDs, id)
	}
	rows.Close()

	if len(resourceIDs) == 0 {
		return &noResourcesFoundError{}
	}

	// commit current transaction and start a new one
	err = tx.Commit()
	if err != nil {
		return err
	}

	for _, id := range resourceIDs {

		// begin new transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// delete resource rows (not commited yet)
		_, err = tx.ExecContext(
			ctx,
			"DELETE FROM packResources WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND class = $4 AND ready = FALSE",
			*packID, id.roleID, id.stringID, id.class,
		)
		if err != nil {
			return err
		}

		key := resourceKey(*packID, id.roleID, id.stringID, id.class)
		_, err = s3c.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: bucket,
			Key:    &key,
		})
		if err != nil {
			return err
		}

		// commit current transaction
		err = tx.Commit()
		if err != nil {
			return err
		}

	}

	// no error == success :)
	return nil

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
