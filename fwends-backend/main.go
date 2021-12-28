package main

import (
	"fwends-backend/api"
	"fwends-backend/connections"
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

	router := httprouter.New()
	router.POST("/api/auth", api.Authenticate(db, rdb))
	router.GET("/api/auth/config", api.AuthConfig())

	log.Info("Starting http server")
	log.Fatal(http.ListenAndServe(":80", router))
}
