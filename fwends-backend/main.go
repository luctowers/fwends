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

	podIndex, err := util.PodIndex()
	if err != nil {
		log.WithError(err).Fatal("Failed to determine pod index")
	}
	snowflake, err := util.NewSnowflakeGenerator(podIndex)
	if err != nil {
		log.WithError(err).Fatal("Failed to create snowflake id generator")
	}

	router := httprouter.New()
	router.POST("/api/auth", api.Authenticate(db, rdb))
	router.GET("/api/auth", api.AuthVerify(rdb))
	router.GET("/api/auth/config", api.AuthConfig())
	router.POST("/api/packs/", api.CreatePack(db, snowflake))
	router.PUT("/api/packs/:id", api.UpdatePack(db))
	router.DELETE("/api/packs/:id", api.DeletePack(db))

	log.WithFields(log.Fields{
		"podIndex": podIndex,
	}).Info("Starting http server")
	log.Fatal(http.ListenAndServe(":80", router))
}
