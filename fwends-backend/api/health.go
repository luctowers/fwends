package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fwends-backend/config"
	"fwends-backend/util"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// GET /api/health
//
// Gets the health of the services that the backend depends on.
func HealthCheck(cfg *config.Config, logger *zap.Logger, db *sql.DB, rdb *redis.Client, s3c *s3.Client) httprouter.Handle {
	type serviceInfo struct {
		Postgres bool `json:"postgres"`
		Redis    bool `json:"redis"`
		S3       bool `json:"s3"`
	}
	type responseBody struct {
		Services serviceInfo `json:"services"`
	}
	return util.WrapDecoratedHandle(
		cfg, logger,
		func(w http.ResponseWriter, r *http.Request, _ httprouter.Params, logger *zap.Logger) (int, error) {

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
					logger.With(zap.Error(err)).Error("Failed to ping postgres")
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
					logger.With(zap.Error(err)).Error("Failed to ping redis")
				}
				crdb <- err == nil
			}()

			// check for bucket existance in s3
			cs3c := make(chan bool)
			go func() {
				defer wg.Done()
				_, err := s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
				if err != nil {
					logger.With(zap.Error(err)).Error("Failed to list s3 buckets")
				}
				cs3c <- err == nil
			}()

			var resbody responseBody
		loop:
			for {
				select {
				// status received from service
				case resbody.Services.Postgres = <-cdb:
				case resbody.Services.Redis = <-crdb:
				case resbody.Services.S3 = <-cs3c:
				// all service statuses rececived or timed out
				case <-ctx.Done():
					json.NewEncoder(w).Encode(resbody)
					break loop
				// request cancelled
				case <-r.Context().Done():
					break loop
				}
			}

			return http.StatusOK, nil

		},
	)
}
