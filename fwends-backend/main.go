package main

import (
	"fwends-backend/api"
	"fwends-backend/connections"
	"fwends-backend/util"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {

	// bind environment vairables to viper configuration
	viper.AllowEmptyEnv(false)
	viper.BindEnv("log_debug")
	viper.BindEnv("http_port")
	viper.BindEnv("http_debug")
	viper.BindEnv("auth_enable")
	viper.BindEnv("session_id_size")
	viper.BindEnv("session_ttl")
	viper.BindEnv("session_cookie")
	viper.BindEnv("session_redis_prefix")
	viper.BindEnv("google_client_id")
	viper.BindEnv("postgres_endpoint")
	viper.BindEnv("postgres_user")
	viper.BindEnv("postgres_password")
	viper.BindEnv("postgres_db")
	viper.BindEnv("postgres_ssl_mode")
	viper.BindEnv("redis_endpoint")
	viper.BindEnv("redis_password")
	viper.BindEnv("s3_endpoint")
	viper.BindEnv("s3_region")
	viper.BindEnv("s3_access_key")
	viper.BindEnv("s3_secret_key")
	viper.BindEnv("s3_media_bucket")

	// set config defaults
	viper.SetDefault("log_debug", false)
	viper.SetDefault("http_port", 80)
	viper.SetDefault("http_debug", false)
	viper.SetDefault("auth_enable", true)
	viper.SetDefault("session_id_size", 32)
	viper.SetDefault("session_ttl", 24*time.Hour)
	viper.SetDefault("session_cookie", "fwends_session")
	viper.SetDefault("session_redis_prefix", "session/")
	viper.SetDefault("postgres_ssl_mode", "require")

	// setup logging
	loggerConfig := zap.Config{
		DisableStacktrace: true,
		DisableCaller:     true,
		Development:       viper.GetBool("log_debug"),
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        zapcore.OmitKey,
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	if viper.GetBool("log_debug") {
		loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	logger, _ := loggerConfig.Build()

	// open database connections
	db, err := connections.OpenPostgres()
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create postgres client")
	}
	rdb := connections.OpenRedis()
	s3c, err := connections.OpenS3()
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create s3 client")
	}

	// initialize id generator
	podIndex, err := util.PodIndex()
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to determine pod index")
	}
	snowflake, err := util.NewSnowflakeGenerator(podIndex)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Failed to create snowflake id generator")
	}

	// register http routes
	router := httprouter.New()
	router.GET("/api/health", api.HealthCheck(logger, db, rdb, s3c))
	router.POST("/api/auth", api.Authenticate(logger, db, rdb))
	router.GET("/api/auth", api.AuthVerify(logger, rdb))
	router.GET("/api/auth/config", api.AuthConfig(logger))
	router.POST("/api/packs/", api.CreatePack(logger, db, snowflake))
	router.GET("/api/packs/:pack_id", api.GetPack(logger, db))
	router.PUT("/api/packs/:pack_id", api.UpdatePack(logger, db))
	router.DELETE("/api/packs/:pack_id", api.DeletePack(logger, db, s3c))
	router.DELETE("/api/packs/:pack_id/:role_id", api.DeletePackRole(logger, db, s3c))
	router.DELETE("/api/packs/:pack_id/:role_id/:string_id", api.DeletePackString(logger, db, s3c))
	router.PUT("/api/packs/:pack_id/:role_id/:string_id", api.UploadPackResource(logger, db, s3c))

	// start the server
	logger.With(zap.Int64("podIndex", podIndex)).Info("Starting http server")
	err = http.ListenAndServe(":"+viper.GetString("http_port"), router)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("failed to serve http")
	}

}
