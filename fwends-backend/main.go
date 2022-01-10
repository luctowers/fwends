package main

import (
	"fwends-backend/api"
	"fwends-backend/connections"
	"fwends-backend/util"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {

	// bind environment vairables to viper configuration
	viper.AllowEmptyEnv(false)
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
	viper.SetDefault("http_port", 80)
	viper.SetDefault("http_debug", false)
	viper.SetDefault("auth_enable", true)
	viper.SetDefault("session_id_size", 32)
	viper.SetDefault("session_ttl", 24*time.Hour)
	viper.SetDefault("session_cookie", "fwends_session")
	viper.SetDefault("session_redis_prefix", "session/")
	viper.SetDefault("postgres_ssl_mode", "require")

	// open database connections
	db, err := connections.OpenPostgres()
	if err != nil {
		log.WithError(err).Fatal("Failed to create postgres client")
	}
	rdb := connections.OpenRedis()
	s3c, err := connections.OpenS3()
	if err != nil {
		log.WithError(err).Fatal("Failed to create s3 client")
	}

	// initialize id generator
	podIndex, err := util.PodIndex()
	if err != nil {
		log.WithError(err).Fatal("Failed to determine pod index")
	}
	snowflake, err := util.NewSnowflakeGenerator(podIndex)
	if err != nil {
		log.WithError(err).Fatal("Failed to create snowflake id generator")
	}

	// register http routes
	router := httprouter.New()
	router.GET("/api/health", api.HealthCheck(db, rdb, s3c))
	router.POST("/api/auth", api.Authenticate(db, rdb))
	router.GET("/api/auth", api.AuthVerify(rdb))
	router.GET("/api/auth/config", api.AuthConfig())
	router.POST("/api/packs/", api.CreatePack(db, snowflake))
	router.GET("/api/packs/:pack_id", api.GetPack(db))
	router.PUT("/api/packs/:pack_id", api.UpdatePack(db))
	router.DELETE("/api/packs/:pack_id", api.DeletePack(db, s3c))
	router.DELETE("/api/packs/:pack_id/:role_id", api.DeletePackRole(db, s3c))
	router.DELETE("/api/packs/:pack_id/:role_id/:string_id", api.DeletePackString(db, s3c))
	router.PUT("/api/packs/:pack_id/:role_id/:string_id", api.UploadPackResource(db, s3c))

	// start the server
	log.WithFields(log.Fields{
		"podIndex": podIndex,
	}).Info("Starting http server")
	log.Fatal(http.ListenAndServe(":"+viper.GetString("http_port"), router))

}
