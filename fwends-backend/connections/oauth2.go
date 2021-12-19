package connections

import (
	"net/http"

	"google.golang.org/api/oauth2/v2"
)

func OpenOauth2() (*oauth2.Service, error) {
	var httpClient = &http.Client{}
	return oauth2.New(httpClient)
}
