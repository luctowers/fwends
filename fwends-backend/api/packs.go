package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"fwends-backend/config"
	"fwends-backend/util"
	"net/http"
	"regexp"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/julienschmidt/httprouter"
	"github.com/lib/pq"
	"go.uber.org/zap"
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

// POST /api/packs/
//
// Creates an new pack with a title and returns the id.
func CreatePack(cfg *config.Config, logger *zap.Logger, db *sql.DB, snowflake *util.SnowflakeGenerator) httprouter.Handle {
	type requestBody struct {
		Title string `json:"title"`
	}
	type responseBody struct {
		// encode as string because javascript doesn't play nice with 64-bit ints
		ID int64 `json:"id,string"`
	}
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params, logger *zap.Logger) (int, error) {

			// decode request body
			decoder := json.NewDecoder(r.Body)
			var reqbody requestBody
			err := decoder.Decode(&reqbody)
			if err != nil {
				return http.StatusBadRequest, fmt.Errorf("failed to decode resonse body: %v", err)
			} else if len(reqbody.Title) == 0 {
				return http.StatusBadRequest, errors.New("empty pack title is not allowed")
			}

			// this id will be used for the life of pack
			id := snowflake.GenID()

			// insert pack row in postgres
			_, err = db.ExecContext(
				r.Context(), "INSERT INTO packs (id, title) VALUES ($1, $2)",
				id, reqbody.Title,
			)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// respond to reqquest
			w.Header().Set("Content-Type", "application/json")
			var resbody responseBody
			resbody.ID = id
			json.NewEncoder(w).Encode(resbody)

			return http.StatusOK, nil

		},
	)
}

// GET /api/packs/:pack_id
//
// Gets a pack's title.
func GetPack(cfg *config.Config, logger *zap.Logger, db *sql.DB) httprouter.Handle {
	type packString struct {
		HasAudio bool `json:"audio"`
		HasImage bool `json:"image"`
	}
	type responseBody struct {
		Title     string                           `json:"title"`
		Resources map[string]map[string]packString `json:"resources"`
	}
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {

			// get path params
			id := ps.ByName("pack_id")

			var resbody responseBody

			// query postgres for pack title
			rows, err := db.QueryContext(r.Context(), "SELECT title FROM packs WHERE id = $1", id)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			defer rows.Close()
			if !rows.Next() {
				// no row was returned
				return http.StatusNotFound, errors.New("pack not found")
			}
			err = rows.Scan(&resbody.Title)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			rows.Close()

			// query postgres for pack resources
			rows, err = db.QueryContext(r.Context(), "SELECT roleId, stringId, class FROM packResources WHERE packId = $1 AND ready = TRUE", id)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			defer rows.Close()
			resbody.Resources = make(map[string]map[string]packString)
			for rows.Next() {
				var roleID string
				var stringID string
				var resourceClass string
				rows.Scan(&roleID, &stringID, &resourceClass)
				roleMap, rolePresent := resbody.Resources[roleID]
				if !rolePresent {
					roleMap = make(map[string]packString)
					resbody.Resources[roleID] = roleMap
				}
				str, strPresent := roleMap[stringID]
				if !strPresent {
					str = packString{}
				}
				switch resourceClass {
				case "audio":
					str.HasAudio = true
				case "image":
					str.HasImage = true
				default:
					logger.Warn("unknown pack resource class found in database", zap.String("resourceClass", resourceClass))
				}
				roleMap[stringID] = str
			}
			rows.Close()

			// respond to request
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resbody)

			return http.StatusOK, nil

		},
	)
}

// PUT /api/packs/:pack_id
//
// Updates a pack's title.
func UpdatePack(cfg *config.Config, logger *zap.Logger, db *sql.DB) httprouter.Handle {
	type requestBody struct {
		Title string `json:"title"`
	}
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {

			// get path params
			id := ps.ByName("pack_id")

			// decode request body
			decoder := json.NewDecoder(r.Body)
			var reqbody requestBody
			err := decoder.Decode(&reqbody)
			if err != nil {
				return http.StatusBadRequest, fmt.Errorf("failed to decode resonse body: %v", err)
			} else if len(reqbody.Title) == 0 {
				return http.StatusBadRequest, errors.New("empty pack title is not allowed")
			}

			// update pack title
			res, err := db.ExecContext(r.Context(), "UPDATE packs SET title = $2 WHERE id = $1", id, reqbody.Title)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// check whether anything was updated
			affected, err := res.RowsAffected()
			if err != nil {
				return http.StatusInternalServerError, err
			} else if affected != 1 {
				// row was not changed, the pack does not exist
				return http.StatusNotFound, errors.New("pack not found")
			}

			return http.StatusOK, nil

		},
	)
}

// PUT /api/packs/:pack_id/:role_id/:string_id
//
// Adds or replaces a image or audio pack resource.
func UploadPackResource(cfg *config.Config, logger *zap.Logger, db *sql.DB, s3c *s3.Client) httprouter.Handle {
	identifierExpression := regexp.MustCompile(`^[a-z0-9_]{1,63}$`)
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {

			// get path params
			packID := ps.ByName("pack_id")
			roleID := ps.ByName("role_id")
			stringID := ps.ByName("string_id")

			// validation
			if !identifierExpression.MatchString(roleID) {
				return http.StatusBadRequest, fmt.Errorf("failed to validate role id: %v", roleID)
			}
			if !identifierExpression.MatchString(stringID) {
				return http.StatusBadRequest, fmt.Errorf("failed to validate string id: %v", stringID)
			}

			// determine whether upload is audio or image
			contentType := r.Header.Get("Content-Type")
			resourceClass, err := deriveResourceClass(contentType)
			if err != nil {
				return http.StatusBadRequest, err
			}

			// begin a new transcation
			tx, err := db.BeginTx(r.Context(), nil)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			defer tx.Rollback()

			// check whether resource already exists, and wether is ready state
			rows, err := tx.QueryContext(
				r.Context(),
				"SELECT ready FROM packResources WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND class = $4",
				packID, roleID, stringID, resourceClass,
			)
			if err != nil {
				return http.StatusInternalServerError, err
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
					return http.StatusNotFound, errors.New("pack not found")
				} else if err != nil {
					return http.StatusInternalServerError, err
				}

			} else if resourceReady {

				// update existing resource to be not ready
				result, err := tx.ExecContext(
					r.Context(),
					"UPDATE packResources SET ready = FALSE WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND class = $4",
					packID, roleID, stringID, resourceClass,
				)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				rowsAffected, _ := result.RowsAffected()
				if rowsAffected != 1 {
					return http.StatusInternalServerError, errors.New("failed to mark pack resource as not ready")
				}

			}

			// commit current transaction and start a new one
			err = tx.Commit()
			if err != nil {
				return http.StatusInternalServerError, err
			}
			tx, err = db.BeginTx(r.Context(), nil)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			defer tx.Rollback()

			// indicate resource is ready (not commited yet), goal is to lock the row
			result, err := tx.ExecContext(
				r.Context(),
				"UPDATE packResources SET ready = TRUE WHERE packId = $1 AND roleId = $2 AND stringId = $3 AND class = $4 AND ready = FALSE",
				packID, roleID, stringID, resourceClass,
			)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected != 1 {
				return http.StatusInternalServerError, errors.New("failed to mark pack resource as ready")
			}

			// upload resource, now that the row is lock
			key := resourceKey(packID, roleID, stringID, resourceClass)
			_, err = s3c.PutObject(r.Context(), &s3.PutObjectInput{
				Bucket:        &cfg.S3.MediaBucket,
				Key:           &key,
				Body:          r.Body,
				ContentLength: r.ContentLength,
				ContentType:   &contentType,
			}, s3.WithAPIOptions(
				v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
			))
			if err != nil {
				return http.StatusInternalServerError, err
			} else {
				logger.With(
					zap.String("key", key),
				).Info("uploaded new pack resource")
			}

			// commit current transaction
			err = tx.Commit()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			return http.StatusOK, nil

		},
	)
}

// DELETE /api/packs/:pack_id
//
// Deletes a pack and its ascociated resources.
func DeletePack(cfg *config.Config, logger *zap.Logger, db *sql.DB, s3c *s3.Client) httprouter.Handle {
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {

			// get path params
			packID := ps.ByName("pack_id")

			// delete pack resources
			err := deleteRelatedResources(r.Context(), db, s3c, &cfg.S3.MediaBucket, &packID, nil, nil)
			if err != nil {
				if err != nil {
					switch err.(type) {
					case *noResourcesFoundError:
						// a pack is allowed to have no resources
					default:
						return http.StatusInternalServerError, fmt.Errorf("failed to delete pack resource: %v", err)
					}
				}
			}

			// delete pack
			res, err := db.ExecContext(r.Context(), "DELETE FROM packs WHERE id = $1;", packID)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// check how many rows were affected
			affected, err := res.RowsAffected()
			if err != nil {
				return http.StatusInternalServerError, err
			} else if affected != 1 {
				return http.StatusNotFound, errors.New("pack not found")
			}

			return http.StatusOK, nil

		},
	)
}

// DELETE /api/packs/:pack_id/:role_id
//
// Deletes all pack resources belonging to a role.
func DeletePackRole(cfg *config.Config, logger *zap.Logger, db *sql.DB, s3c *s3.Client) httprouter.Handle {
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {

			// get path params
			packID := ps.ByName("pack_id")
			roleID := ps.ByName("role_id")

			// delete resources
			err := deleteRelatedResources(r.Context(), db, s3c, &cfg.S3.MediaBucket, &packID, &roleID, nil)
			if err != nil {
				if err != nil {
					switch err.(type) {
					case *noResourcesFoundError:
						return http.StatusNotFound, errors.New("no resources found in pack role")
					default:
						return http.StatusInternalServerError, fmt.Errorf("failed to delete pack role resource: %v", err)
					}
				}
			}

			return http.StatusOK, nil

		},
	)
}

// PUT /api/packs/:pack_id/:role_id/:string_id
//
// Deletes all pack resources belonging to a string.
func DeletePackString(cfg *config.Config, logger *zap.Logger, db *sql.DB, s3c *s3.Client) httprouter.Handle {
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *zap.Logger) (int, error) {

			// get path params
			packID := ps.ByName("pack_id")
			roleID := ps.ByName("role_id")
			stringID := ps.ByName("string_id")

			// delete resources
			err := deleteRelatedResources(r.Context(), db, s3c, &cfg.S3.MediaBucket, &packID, &roleID, &stringID)
			if err != nil {
				if err != nil {
					switch err.(type) {
					case *noResourcesFoundError:
						return http.StatusNotFound, errors.New("no resources found in pack string")
					default:
						return http.StatusInternalServerError, fmt.Errorf("failed to delete pack string resource: %v", err)
					}
				}
			}

			return http.StatusOK, nil

		},
	)
}

// HELPERS

// returned by deleteRelatedResources when do resources are found to delete
type noResourcesFoundError struct{}

func (e *noResourcesFoundError) Error() string {
	return "no pack resources found"
}

// Shared functionality for all pack delete operations
func deleteRelatedResources(ctx context.Context, db *sql.DB, s3c *s3.Client, bucket *string, packID *string, roleID *string, stringID *string) error {

	// each mode corresponds one of the above http DELETE specificities
	const (
		deletePackMode   = 0
		deleteRoleMode   = 1
		deleteStringMode = 2
	)

	// detect mode
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

		// delete the object from s3
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

// Generates a s3 key for a resource.
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
	// determine resource type from extension
	switch contentType {

	// image content types
	case "image/webp":
		fallthrough
	case "image/jpeg":
		fallthrough
	case "image/png":
		fallthrough
	case "image/svg+xml":
		return "image", nil

	// audio content types
	case "audio/aac":
		fallthrough
	case "audio/mpeg":
		fallthrough
	case "audio/wav":
		fallthrough
	case "audio/flac":
		return "audio", nil

	// unknown content types
	default:
		return "", errors.New("unsupported pack resource content type")
	}
}
