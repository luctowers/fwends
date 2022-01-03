package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

type healthInfo struct {
	Services healthServiceInfo `json:"services"`
}

type healthServiceInfo struct {
	Postgres bool `json:"postgres"`
	Redis    bool `json:"redis"`
}

func HealthCheck(db *sql.DB, rdb *redis.Client) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		cdb := make(chan error)
		crdb := make(chan *redis.StatusCmd)
		go func() {
			cdb <- db.Ping()
		}()
		go func() {
			crdb <- rdb.Ping(context.Background())
		}()
		var health healthInfo
		err := <-cdb
		health.Services.Postgres = err == nil
		if err != nil {
			log.WithError(err).Error("Unable to ping postgres")
		}
		_, err = (<-crdb).Result()
		health.Services.Redis = err == nil
		if err != nil {
			log.WithError(err).Error("Unable to ping redis")
		}
		json.NewEncoder(w).Encode(health)
	}
}
