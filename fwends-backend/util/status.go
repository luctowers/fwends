package util

import (
	"fmt"
	"net/http"
)

func Error(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func OK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, http.StatusText(http.StatusOK))
}
