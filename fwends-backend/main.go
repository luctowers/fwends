package main

import (
	"fmt"
	"fwends-backend/api"
	"fwends-backend/config"
	"fwends-backend/services"
	"fwends-backend/util"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {

	// load config
	cfg := &config.Config{}
	v := viper.New()
	config.BindEnv(v)
	config.SetDefaults(v)
	err := v.UnmarshalExact(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// validate config
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// create zap logger
	logger, err := services.NewLogger(cfg.LogDebug)
	if err != nil {
		// this is the final usage of the default go logger
		log.Fatal(err)
	}

	// open service connections
	db, err := services.NewPostgres(&cfg.Postgres)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to create postgres client")
	}
	rdb := services.NewRedis(&cfg.Redis)
	s3c, err := services.NewS3(&cfg.S3)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to create s3 client")
	}

	// initialize id generator
	podIndex, err := util.PodIndex()
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to determine pod index")
	}
	snowflake, err := util.NewSnowflakeGenerator(podIndex)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to create snowflake id generator")
	}

	// register http routes
	router := httprouter.New()
	router.GET("/api/health", api.HealthCheck(cfg, logger, db, rdb, s3c))
	router.POST("/api/auth", api.Authenticate(cfg, logger, db, rdb))
	router.GET("/api/auth", api.AuthVerify(cfg, logger, rdb))
	router.GET("/api/auth/config", api.AuthConfig(cfg, logger))
	router.POST("/api/packs/", api.CreatePack(cfg, logger, db, snowflake))
	router.GET("/api/packs/:pack_id", api.GetPack(cfg, logger, db))
	router.PUT("/api/packs/:pack_id", api.UpdatePack(cfg, logger, db))
	router.DELETE("/api/packs/:pack_id", api.DeletePack(cfg, logger, db, s3c))
	router.DELETE("/api/packs/:pack_id/:role_id", api.DeletePackRole(cfg, logger, db, s3c))
	router.DELETE("/api/packs/:pack_id/:role_id/:string_id", api.DeletePackString(cfg, logger, db, s3c))
	router.PUT("/api/packs/:pack_id/:role_id/:string_id", api.UploadPackResource(cfg, logger, db, s3c))

	// start the server
	logger.With(zap.Int64("podIndex", podIndex)).Info("starting http server")
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	err = http.ListenAndServe(addr, router)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to serve http")
	}

}
