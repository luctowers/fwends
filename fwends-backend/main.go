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
	oauth2, err := connections.OpenOauth2()
	if err != nil {
		log.WithError(err).Fatal("Failed to create oauth2 client")
	}

	router := httprouter.New()
	router.POST("/api/authenticate", api.Authenticate(db, rdb, oauth2))

	log.Info("Starting http server")
	log.Fatal(http.ListenAndServe(":80", router))
}
