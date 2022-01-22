package main

import (
	"database/sql"
	"fmt"
	"fwends-backend/api"
	"fwends-backend/config"
	"fwends-backend/handler"
	"fwends-backend/services"
	"fwends-backend/util"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// the following functions are defined below and will panic if not successful
	cfg := loadConfig()
	logger := newLogger(cfg)
	db := newPostgres(cfg)
	rdb := newRedis(cfg)
	s3c := newS3(cfg)
	podIndex := getPodIndex()
	idgen := newIDGenerator(podIndex)

	// wrapper for handlers
	w := func(h handler.Handler) httprouter.Handle {
		h = handler.NewLoggingHandler(h)
		h = handler.NewStatusHandler(h, cfg.HTTPDebug)
		return handler.ToHTTPRouterHandle(h, logger)
	}

	// register http routes
	router := httprouter.New()
	router.GET("/api/health", w(api.HealthCheck(cfg, db, rdb, s3c)))
	router.POST("/api/auth", w(api.Authenticate(cfg, db, rdb)))
	router.GET("/api/auth", w(api.AuthVerify(cfg, rdb)))
	router.GET("/api/auth/config", w(api.AuthConfig(cfg)))
	router.POST("/api/packs/", w(api.CreatePack(db, idgen)))
	router.PUT("/api/packs/:pack_id", w(api.UpdatePack(db)))
	router.GET("/api/packs/", w(api.ListPacks(cfg, db)))
	router.GET("/api/packs/:pack_id", w(api.GetPack(cfg, db)))
	router.DELETE("/api/packs/:pack_id", w(api.DeletePack(cfg, db, s3c)))
	router.DELETE("/api/packs/:pack_id/:role_id", w(api.DeletePackRole(cfg, db, s3c)))
	router.DELETE("/api/packs/:pack_id/:role_id/:string_id", w(api.DeletePackString(cfg, db, s3c)))
	router.PUT("/api/packs/:pack_id/:role_id/:string_id", w(api.UploadPackResource(cfg, db, s3c, idgen)))

	// start the server
	logger.With(zap.Int64("podIndex", podIndex)).Info("starting http server")
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to serve http")
	}
}

func loadConfig() *config.Config {
	cfg := &config.Config{}
	v := viper.New()
	config.BindEnv(v)
	config.SetDefaults(v)
	err := v.UnmarshalExact(cfg)
	if err != nil {
		panic(err)
	}
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

func newLogger(cfg *config.Config) *zap.Logger {
	logger, err := services.NewLogger(cfg.LogDebug)
	if err != nil {
		panic(err)
	}
	return logger
}

func newPostgres(cfg *config.Config) *sql.DB {
	db, err := services.NewPostgres(&cfg.Postgres)
	if err != nil {
		panic(err)
	}
	return db
}

func newRedis(cfg *config.Config) *redis.Client {
	return services.NewRedis(&cfg.Redis)
}

func newS3(cfg *config.Config) *s3.Client {
	s3c, err := services.NewS3(&cfg.S3)
	if err != nil {
		panic(err)
	}
	return s3c
}

func getPodIndex() int64 {
	podIndex, err := util.PodIndex()
	if err != nil {
		panic(err)
	}
	return podIndex
}

func newIDGenerator(podIndex int64) *util.SnowflakeGenerator {
	snowflake, err := util.NewSnowflakeGenerator(podIndex)
	if err != nil {
		panic(err)
	}
	return snowflake
}
