package api

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"fwends-backend/config"
	"fwends-backend/handler"
	"fwends-backend/util"
	"net/http"
	"regexp"
	"strconv"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/lib/pq"
)

/*
Example curl commands:

curl -X GET http://localhost:8080/api/packs/
curl -X POST http://localhost:8080/api/packs/ -d '{"title":"Test Pack"}'
curl -X GET http://localhost:8080/api/packs/6882582496895041536
curl -X PUT http://localhost:8080/api/packs/6882582496895041536 -d '{"title":"Updated Test Pack"}'
curl -X PUT http://localhost:8080/api/packs/6882582496895041536/role/string -H 'Content-Type: image/png' --data-binary "@path/to/image.png"
curl -X PUT http://localhost:8080/api/packs/6882582496895041536/role/string -H 'Content-Type: audio/mpeg' --data-binary "@path/to/audio.mp3"
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536/role
curl -X DELETE http://localhost:8080/api/packs/6882582496895041536/role/string
*/

// GET /api/packs/
//
// Lists all existing packs.
func ListPacks(cfg *config.Config, db *sql.DB) handler.Handler {
	return &listPacksHandler{cfg, db}
}

type listPacksHandler struct {
	cfg *config.Config
	db  *sql.DB
}

func (h *listPacksHandler) Handle(i handler.Input) (int, error) {
	packs := make([]packSummary, 0)

	rows, err := h.db.QueryContext(i.Request.Context(), `
		SELECT
			packs.pack_id,
			packs.title,
			packs.hash,
			COUNT(DISTINCT pack_resources.role_id),
			COUNT(DISTINCT pack_resources.role_id || '-' || pack_resources.string_id)
		FROM packs
			LEFT OUTER JOIN pack_resources ON pack_resources.pack_id = packs.pack_id
		GROUP BY packs.pack_id
	`)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer rows.Close()
	for rows.Next() {
		var pack packSummary
		var hash []byte
		err := rows.Scan(&pack.ID, &pack.Title, &hash, &pack.RoleCount, &pack.StringCount)
		pack.Hash = hex.EncodeToString(hash)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		packs = append(packs, pack)
	}
	rows.Close()

	// respond to request
	i.Response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(i.Response).Encode(packs)

	return http.StatusOK, nil
}

// POST /api/packs/
//
// Creates an new pack with a title and returns the id.
func CreatePack(db *sql.DB, idgen *util.SnowflakeGenerator) handler.Handler {
	return &createPackHandler{db, idgen}
}

type createPackHandler struct {
	db    *sql.DB
	idgen *util.SnowflakeGenerator
}

func (h *createPackHandler) Handle(i handler.Input) (int, error) {
	// decode request body
	decoder := json.NewDecoder(i.Request.Body)
	var reqbody struct {
		Title string `json:"title"`
	}
	err := decoder.Decode(&reqbody)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to decode resonse body: %v", err)
	} else if len(reqbody.Title) == 0 {
		return http.StatusBadRequest, errors.New("empty pack title is not allowed")
	}

	// this id will be used for the life of pack
	id := h.idgen.GenID()

	// insert pack row in postgres
	_, err = h.db.ExecContext(i.Request.Context(),
		"INSERT INTO packs (pack_id, title) VALUES ($1, $2)",
		id, reqbody.Title,
	)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// respond to request
	var resbody struct {
		// encode as string because javascript doesn't play nice with 64-bit ints
		ID   int64  `json:"id,string"`
		Hash string `json:"hash"`
	}
	resbody.ID = id
	// sha256 of nothing
	resbody.Hash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	i.Response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(i.Response).Encode(resbody)

	return http.StatusOK, nil
}

// GET /api/packs/:pack_id
//
// Gets a pack's title.
func GetPack(cfg *config.Config, db *sql.DB) handler.Handler {
	return &getPackHandler{cfg, db}
}

type getPackHandler struct {
	cfg *config.Config
	db  *sql.DB
}

func (h *getPackHandler) Handle(i handler.Input) (int, error) {
	packID := i.Params.ByName("pack_id")

	var resbody struct {
		Title string     `json:"title"`
		Hash  string     `json:"hash"`
		Roles []packRole `json:"roles"`
	}

	// start a new transaction to ensure consistent state
	tx, err := h.db.BeginTx(i.Request.Context(), &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer tx.Rollback()

	// query postgres for pack title
	resbody.Title, resbody.Hash, err = h.getPackDetails(i.Request.Context(), tx, packID)
	if err == errPackNotFoundError {
		return http.StatusNotFound, nil
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	// get pack resources
	resbody.Roles, err = h.getPackResources(i.Request.Context(), tx, packID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// respond to request
	i.Response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(i.Response).Encode(resbody)

	return http.StatusOK, nil
}

func (h *getPackHandler) getPackDetails(ctx context.Context, tx *sql.Tx, packID string) (string, string, error) {
	rows, err := tx.QueryContext(ctx,
		"SELECT title, hash FROM packs WHERE pack_id = $1", packID,
	)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()
	if !rows.Next() {
		// no row was returned
		return "", "", errPackNotFoundError
	}
	var title string
	var hash []byte
	err = rows.Scan(&title, &hash)
	if err != nil {
		return "", "", err
	}
	return title, hex.EncodeToString(hash), nil
}

func (h *getPackHandler) getPackResources(ctx context.Context, tx *sql.Tx, packID string) ([]packRole, error) {
	// query postgres for pack resources
	rows, err := tx.QueryContext(ctx,
		`
		SELECT role_id, string_id, resource_class, resource_id
		FROM pack_resources WHERE pack_id = $1
		ORDER BY role_id, string_id
		`,
		packID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]packRole, 0)
	var prevRoleID string
	var prevStringID string
	for rows.Next() {
		var roleID string
		var stringID string
		var resourceClass string
		var resourceID int64
		err := rows.Scan(&roleID, &stringID, &resourceClass, &resourceID)
		if err != nil {
			return nil, err
		}
		if roleID != prevRoleID {
			roles = append(roles, packRole{
				ID:      roleID,
				Strings: []packString{{ID: stringID}},
			})
		} else if stringID != prevStringID {
			role := &roles[len(roles)-1]
			role.Strings = append(role.Strings, packString{ID: stringID})
		}
		role := &roles[len(roles)-1]
		str := &role.Strings[len(role.Strings)-1]
		switch resourceClass {
		case "audio":
			str.Audio = resourceID
		case "image":
			str.Image = resourceID
		}
		prevRoleID = roleID
		prevStringID = stringID
	}
	return roles, nil
}

// PUT /api/packs/:pack_id
//
// Updates a pack's title.
func UpdatePack(db *sql.DB) handler.Handler {
	return &updatePackHandler{db}
}

type updatePackHandler struct {
	db *sql.DB
}

func (h *updatePackHandler) Handle(i handler.Input) (int, error) {
	packID := i.Params.ByName("pack_id")

	// decode request body
	decoder := json.NewDecoder(i.Request.Body)
	var reqbody struct {
		Title string `json:"title"`
	}
	err := decoder.Decode(&reqbody)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("failed to decode resonse body: %v", err)
	} else if len(reqbody.Title) == 0 {
		return http.StatusBadRequest, errors.New("empty pack title is not allowed")
	}

	// update pack title
	res, err := h.db.ExecContext(i.Request.Context(),
		"UPDATE packs SET title = $2 WHERE pack_id = $1", packID, reqbody.Title,
	)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// check whether anything was updated
	affected, err := res.RowsAffected()
	if err != nil {
		return http.StatusInternalServerError, err
	} else if affected != 1 {
		// row was not changed, the pack does not exist
		return http.StatusNotFound, nil
	}

	return http.StatusOK, nil
}

// PUT /api/packs/:pack_id/:role_id/:string_id
//
// Adds or replaces a image or audio pack resource.
func UploadPackResource(cfg *config.Config, db *sql.DB, s3c *s3.Client, idgen *util.SnowflakeGenerator) handler.Handler {
	return &uploadPackResourceHandler{idgen, packResourceHandler{cfg, db, s3c}}
}

type uploadPackResourceHandler struct {
	idgen *util.SnowflakeGenerator
	packResourceHandler
}

func (h *uploadPackResourceHandler) Handle(i handler.Input) (int, error) {
	packID := i.Params.ByName("pack_id")
	roleID := i.Params.ByName("role_id")
	stringID := i.Params.ByName("string_id")

	// validation
	if !packResourceIDRegex.MatchString(roleID) {
		return http.StatusBadRequest, fmt.Errorf("failed to validate role id: %v", roleID)
	}
	if !packResourceIDRegex.MatchString(stringID) {
		return http.StatusBadRequest, fmt.Errorf("failed to validate string id: %v", stringID)
	}

	// determine whether upload is audio or image
	contentType := i.Request.Header.Get("Content-Type")
	resourceClass, err := derivePackResourceClass(contentType)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// check whether pack exists
	packExists, err := h.packExists(i.Request.Context(), packID)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !packExists {
		return http.StatusNotFound, nil
	}

	// defer prune the resource, in-case the transaction does not complete
	resourceID := strconv.FormatInt(h.idgen.GenID(), 10)
	transactionCommited := false
	defer func() {
		if !transactionCommited {
			go h.pruneResource(context.Background(), resourceID)
		}
	}()

	// upload the resource to s3
	err = h.uploadResource(i.Request, resourceID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// loop that will retry if transaction serialization anomaly occurs
	for !transactionCommited {
		prevResourceID, err := h.updateResourceIDTransaction(i.Request.Context(),
			packID, roleID, stringID, resourceClass, resourceID,
		)
		if isRetryableSerializationFailure(err) {
			continue
		} else if err == errPackNotFoundError {
			return http.StatusNotFound, err
		} else if err != nil {
			return http.StatusInternalServerError, err
		}
		transactionCommited = true
		if prevResourceID != "" {
			go h.pruneResource(context.Background(), prevResourceID)
		}
	}

	// repond with new resource id
	i.Response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(i.Response).Encode(resourceID)
	return http.StatusOK, nil
}

func (h *uploadPackResourceHandler) packExists(ctx context.Context, packID string) (bool, error) {
	rows, err := h.db.QueryContext(ctx,
		"SELECT 1 FROM packs WHERE pack_id = $1", packID,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (h *uploadPackResourceHandler) uploadResource(r *http.Request, resourceID string) error {
	// insert row to mark possble existance of resource in s3
	_, err := h.db.ExecContext(r.Context(),
		"INSERT INTO resources (resource_id) VALUES ($1)", resourceID,
	)
	if err != nil {
		return err
	}

	// upload resource to s3
	contentType := r.Header.Get("Content-Type")
	_, err = h.s3c.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:        &h.cfg.S3.MediaBucket,
		Key:           &resourceID,
		Body:          r.Body,
		ContentLength: r.ContentLength,
		ContentType:   &contentType,
	}, s3.WithAPIOptions(
		v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
	))
	return err
}

func (h *uploadPackResourceHandler) updateResourceIDTransaction(
	ctx context.Context, packID string, roleID string, stringID string, resourceClass string, resourceID string,
) (string, error) {
	// begin a new transcation
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// check whether pack exists and previous resource id exist
	rows, err := tx.QueryContext(ctx,
		`
		SELECT pack_resources.resource_id
		FROM packs
			LEFT OUTER JOIN pack_resources ON
				pack_resources.pack_id = $1 AND
				pack_resources.role_id = $2 AND
				pack_resources.string_id = $3 AND
				pack_resources.resource_class = $4
		WHERE packs.pack_id = $1
		`,
		packID, roleID, stringID, resourceClass,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var previousResourceID sql.NullInt64
	if !rows.Next() {
		// no row returned, the pack does not exist
		return "", errPackNotFoundError
	} else {
		rows.Scan(&previousResourceID)
	}
	rows.Close()

	// determine whether to use and update or insert statement
	var putResourceStatement string
	if previousResourceID.Valid {
		putResourceStatement = `
			UPDATE pack_resources
				SET resource_id = $5
			WHERE
				pack_id = $1 AND
				role_id = $2 AND
				string_id = $3 AND
				resource_class = $4
		`
	} else {
		putResourceStatement = `
			INSERT INTO pack_resources
				(pack_id, role_id, string_id, resource_class, resource_id)
			VALUES
				($1, $2, $3, $4, $5)
		`
	}

	// put pack_resource row
	result, err := tx.ExecContext(ctx,
		putResourceStatement,
		packID, roleID, stringID, resourceClass, resourceID,
	)
	if err != nil {
		return "", err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		return "", errors.New("inconsistent rows affected")
	}

	// update pack hash if a new row was inserted
	if !previousResourceID.Valid {
		err = h.updatePackHash(ctx, tx, packID)
		if err != nil {
			return "", err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return "", err
	}

	// return previous resource if it exists
	if previousResourceID.Valid {
		return strconv.FormatInt(previousResourceID.Int64, 10), nil
	}

	return "", nil
}

// DELETE /api/packs/:pack_id
//
// Deletes a pack and its ascociated resources.
func DeletePack(cfg *config.Config, db *sql.DB, s3c *s3.Client) handler.Handler {
	return &deletePackHandler{packResourceHandler{cfg, db, s3c}}
}

type deletePackHandler struct {
	packResourceHandler
}

func (h *deletePackHandler) Handle(i handler.Input) (int, error) {
	packID := i.Params.ByName("pack_id")

	for {
		resourcesDeleted, err := h.transaction(i.Request.Context(), packID)
		if err != nil {
			if isRetryableSerializationFailure(err) {
				continue
			}
			return http.StatusInternalServerError, err
		}
		for _, resourceID := range resourcesDeleted {
			go h.pruneResource(context.Background(), resourceID)
		}
		break
	}

	return http.StatusOK, nil
}

func (h *deletePackHandler) transaction(ctx context.Context, packID string) ([]string, error) {
	// begin a new transcation
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// delete resources and get their id for pruning
	rows, err := tx.QueryContext(ctx,
		"DELETE FROM pack_resources WHERE pack_id = $1 RETURNING resource_id",
		packID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resourcesDeleted := make([]string, 0)
	for rows.Next() {
		var resourceID string
		err := rows.Scan(&resourceID)
		if err != nil {
			return nil, err
		}
		resourcesDeleted = append(resourcesDeleted, resourceID)
	}

	// delete pack resources and returns the resource ids
	_, err = tx.ExecContext(ctx,
		"DELETE FROM packs WHERE pack_id = $1", packID,
	)
	if err != nil {
		return nil, err
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return resourcesDeleted, nil
}

// DELETE /api/packs/:pack_id/:role_id
//
// Deletes all pack resources belonging to a role.
func DeletePackRole(cfg *config.Config, db *sql.DB, s3c *s3.Client) handler.Handler {
	return &deletePackRoleHandler{packResourceHandler{cfg, db, s3c}}
}

type deletePackRoleHandler struct {
	packResourceHandler
}

func (h *deletePackRoleHandler) Handle(i handler.Input) (int, error) {
	packID := i.Params.ByName("pack_id")
	roleID := i.Params.ByName("role_id")

	for {
		resourcesDeleted, err := h.transaction(i.Request.Context(), packID, roleID)
		if err != nil {
			if isRetryableSerializationFailure(err) {
				continue
			}
			return http.StatusInternalServerError, err
		}
		for _, resourceID := range resourcesDeleted {
			go h.pruneResource(context.Background(), resourceID)
		}
		break
	}

	return http.StatusOK, nil
}

func (h *deletePackRoleHandler) transaction(ctx context.Context, packID string, roleID string) ([]string, error) {
	// begin a new transcation
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// delete resources and get their id for pruning
	rows, err := tx.QueryContext(ctx,
		"DELETE FROM pack_resources WHERE pack_id = $1 AND role_id = $2 RETURNING resource_id",
		packID, roleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resourcesDeleted := make([]string, 0)
	for rows.Next() {
		var resourceID string
		err := rows.Scan(&resourceID)
		if err == nil {
			resourcesDeleted = append(resourcesDeleted, resourceID)
		}
	}
	rows.Close()

	// recompute the pack hash if anything was deleted
	if len(resourcesDeleted) > 0 {
		err = h.updatePackHash(ctx, tx, packID)
		if err != nil {
			return nil, err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return resourcesDeleted, nil
}

// // PUT /api/packs/:pack_id/:role_id/:string_id
// //
// // Deletes all pack resources belonging to a string.
func DeletePackString(cfg *config.Config, db *sql.DB, s3c *s3.Client) handler.Handler {
	return &deletePackStringHandler{packResourceHandler{cfg, db, s3c}}
}

type deletePackStringHandler struct {
	packResourceHandler
}

func (h *deletePackStringHandler) Handle(i handler.Input) (int, error) {
	packID := i.Params.ByName("pack_id")
	roleID := i.Params.ByName("role_id")
	stringID := i.Params.ByName("string_id")

	for {
		resourcesDeleted, err := h.transaction(i.Request.Context(), packID, roleID, stringID)
		if err != nil {
			if isRetryableSerializationFailure(err) {
				continue
			}
			return http.StatusInternalServerError, err
		}
		for _, resourceID := range resourcesDeleted {
			go h.pruneResource(context.Background(), resourceID)
		}
		break
	}

	return http.StatusOK, nil
}

func (h *deletePackStringHandler) transaction(
	ctx context.Context, packID string, roleID string, stringID string,
) ([]string, error) {
	// begin a new transcation
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// delete resources and get their id for pruning
	rows, err := tx.QueryContext(ctx,
		"DELETE FROM pack_resources WHERE pack_id = $1 AND role_id = $2 and string_id = $3 RETURNING resource_id",
		packID, roleID, stringID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resourcesDeleted := make([]string, 0)
	for rows.Next() {
		var resourceID string
		err := rows.Scan(&resourceID)
		if err == nil {
			resourcesDeleted = append(resourcesDeleted, resourceID)
		}
	}
	rows.Close()

	// recompute the pack hash if anything was deleted
	if len(resourcesDeleted) > 0 {
		err = h.updatePackHash(ctx, tx, packID)
		if err != nil {
			return nil, err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return resourcesDeleted, nil
}

type packResourceHandler struct {
	cfg *config.Config
	db  *sql.DB
	s3c *s3.Client
}

func (h *packResourceHandler) pruneResource(ctx context.Context, id string) error {
	// begin a new transcation
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// atttempt delete row from resources tables
	result, err := tx.ExecContext(ctx, "DELETE FROM resources WHERE resource_id = $1", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// if a row was deleted, move it to pruned_resources and commit and start new transaction
	if rowsAffected == 1 {
		_, err := tx.ExecContext(ctx, "INSERT INTO pruned_resources (resource_id) VALUES ($1)", id)
		if err != nil {
			return err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	// delete from s3
	_, err = h.s3c.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &h.cfg.S3.MediaBucket,
		Key:    &id,
	})
	if err != nil {
		return err
	}

	// attempt delete row from pruned_resources tables
	_, err = h.db.ExecContext(ctx, "DELETE FROM pruned_resources WHERE resource_id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

func (h *packResourceHandler) updatePackHash(ctx context.Context, tx *sql.Tx, packID string) error {
	// query postgres for pack roles and strings
	rows, err := tx.QueryContext(ctx,
		`
		SELECT DISTINCT role_id, string_id
		FROM pack_resources WHERE pack_id = $1
		ORDER BY role_id, string_id
		`,
		packID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	// compute sha256 digest
	hash := sha256.New()
	prevRoleID := ""
	for rows.Next() {
		var roleID string
		var stringID string
		err := rows.Scan(&roleID, &stringID)
		if err != nil {
			return err
		}
		// hash write never returns an error, per https://pkg.go.dev/hash#Hash
		if roleID != prevRoleID {
			_, _ = hash.Write([]byte{0})
			_, _ = hash.Write([]byte(roleID))
		}
		_, _ = hash.Write([]byte{1})
		_, _ = hash.Write([]byte(stringID))
		prevRoleID = roleID
	}
	rows.Close()
	digest := hash.Sum(nil)

	// update the hash
	_, err = tx.ExecContext(ctx,
		"UPDATE packs SET hash = $2 WHERE pack_id = $1", packID, digest,
	)
	if err != nil {
		return err
	}

	return nil
}

//  HELPERS

type packString struct {
	ID    string `json:"id"`
	Audio int64  `json:"audio,string,omitempty"`
	Image int64  `json:"image,string,omitempty"`
}

type packRole struct {
	ID      string       `json:"id"`
	Strings []packString `json:"strings"`
}

type packSummary struct {
	ID          int64  `json:"id,string"`
	Title       string `json:"title"`
	Hash        string `json:"hash"`
	RoleCount   int    `json:"roleCount"`
	StringCount int    `json:"stringCount"`
}

var packResourceIDRegex = regexp.MustCompile(`^[a-z0-9_]{1,63}$`)

var errPackNotFoundError = errors.New("pack not found")

func isRetryableSerializationFailure(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code.Name() == "serialization_failure"
	}
	return false
}

func derivePackResourceClass(contentType string) (string, error) {
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
