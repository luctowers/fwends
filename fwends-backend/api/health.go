package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	S3       bool `json:"s3"`
}

func HealthCheck(db *sql.DB, rdb *redis.Client, s3c *s3.Client) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

		// context that times out after 3 seconds, or when finish is called
		ctx, finish := context.WithTimeout(context.Background(), time.Duration(3*time.Second))
		defer finish()

		// create wait group that will call finish when all services have responded
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer finish()
			wg.Wait()
		}()

		// ping postgres
		cdb := make(chan bool)
		go func() {
			defer wg.Done()
			err := db.PingContext(ctx)
			if err != nil {
				log.WithError(err).Error("Failed to ping postgres")
			}
			cdb <- err == nil
		}()

		// ping redis
		crdb := make(chan bool)
		go func() {
			defer wg.Done()
			cmd := rdb.Ping(ctx)
			_, err := cmd.Result()
			if err != nil {
				log.WithError(err).Error("Failed to ping redis")
			}
			crdb <- err == nil
		}()

		// check for bucket existance in s3
		cs3c := make(chan bool)
		go func() {
			defer wg.Done()
			_, err := s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
			if err != nil {
				log.WithError(err).Error("Failed to list s3 buckets")
			}
			cs3c <- err == nil
		}()

		// select loop
		var health healthInfo
	loop:
		for {
			select {
			case health.Services.Postgres = <-cdb:
			case health.Services.Redis = <-crdb:
			case health.Services.S3 = <-cs3c:
			case <-ctx.Done():
				json.NewEncoder(w).Encode(health)
				break loop
			case <-r.Context().Done():
				// request cancelled
				break loop
			}
		}

	}
}
