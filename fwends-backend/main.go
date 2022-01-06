package main

import (
	"fwends-backend/api"
	"fwends-backend/connections"
	"fwends-backend/util"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
)

func main() {
	db, err := connections.OpenPostgres()
	if err != nil {
		log.WithError(err).Fatal("Failed to create postgres client")
	}
	rdb := connections.OpenRedis()
	s3c, err := connections.OpenS3()
	if err != nil {
		log.WithError(err).Fatal("Failed to create s3 client")
	}

	podIndex, err := util.PodIndex()
	if err != nil {
		log.WithError(err).Fatal("Failed to determine pod index")
	}
	snowflake, err := util.NewSnowflakeGenerator(podIndex)
	if err != nil {
		log.WithError(err).Fatal("Failed to create snowflake id generator")
	}

	router := httprouter.New()
	router.GET("/api/health", api.HealthCheck(db, rdb, s3c))
	router.POST("/api/auth", api.Authenticate(db, rdb))
	router.GET("/api/auth", api.AuthVerify(rdb))
	router.GET("/api/auth/config", api.AuthConfig())
	router.POST("/api/packs/", api.CreatePack(db, snowflake))
	router.GET("/api/packs/:id", api.GetPack(db))
	router.PUT("/api/packs/:id", api.UpdatePack(db))
	router.DELETE("/api/packs/:id", api.DeletePack(db))

	log.WithFields(log.Fields{
		"podIndex": podIndex,
	}).Info("Starting http server")
	log.Fatal(http.ListenAndServe(":80", router))
}
