package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fwends-backend/config"
	"fwends-backend/handler"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// GET /api/health
//
// Gets the health of the services that the backend depends on.
func HealthCheck(cfg *config.Config, db *sql.DB, rdb *redis.Client, s3c *s3.Client) handler.Handler {
	return &healthCheckHandler{cfg, db, rdb, s3c}
}

type healthCheckHandler struct {
	cfg *config.Config
	db  *sql.DB
	rdb *redis.Client
	s3c *s3.Client
}

func (h *healthCheckHandler) Handle(i handler.Input) (int, error) {
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

	// postgres
	cdb := make(chan bool)
	go func() {
		defer wg.Done()
		cdb <- h.checkPostgres(ctx, i.Logger)
	}()

	// redis
	crdb := make(chan bool)
	go func() {
		defer wg.Done()
		crdb <- h.checkRedis(ctx, i.Logger)
	}()

	// s3
	cs3c := make(chan bool)
	go func() {
		defer wg.Done()
		cs3c <- h.checkS3(ctx, i.Logger)
	}()

	// response json
	var resbody struct {
		Services struct {
			Postgres bool `json:"postgres"`
			Redis    bool `json:"redis"`
			S3       bool `json:"s3"`
		} `json:"services"`
	}

	// wait responses and timeouts
loop:
	for {
		select {
		// status received from service
		case resbody.Services.Postgres = <-cdb:
		case resbody.Services.Redis = <-crdb:
		case resbody.Services.S3 = <-cs3c:
		// all service statuses received or timed out
		case <-ctx.Done():
			json.NewEncoder(i.Response).Encode(resbody)
			break loop
		// request cancelled
		case <-i.Request.Context().Done():
			break loop
		}
	}

	return http.StatusOK, nil
}

func (h *healthCheckHandler) checkPostgres(ctx context.Context, logger *zap.Logger) bool {
	if err := h.db.PingContext(ctx); err != nil {
		logger.With(zap.Error(err)).Error("postgres health check failed")
		return false
	} else {
		return true
	}
}

func (h *healthCheckHandler) checkRedis(ctx context.Context, logger *zap.Logger) bool {
	cmd := h.rdb.Ping(ctx)
	if _, err := cmd.Result(); err != nil {
		logger.With(zap.Error(err)).Error("redis health check failed")
		return false
	} else {
		return true
	}
}

func (h *healthCheckHandler) checkS3(ctx context.Context, logger *zap.Logger) bool {
	out, err := h.s3c.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		logger.With(zap.Error(err)).Error("s3 health check failed: failed to list buckets")
		return false
	}
	for _, b := range out.Buckets {
		if *b.Name == h.cfg.S3.MediaBucket {
			return true
		}
	}
	logger.Error("s3 health check failed: media bucket does not exist")
	return false
}
