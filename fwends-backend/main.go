package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type pack struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	RoleCount   int    `json:"roleCount"`
	StringCount int    `json:"stringCount"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api", Hello)
	r.HandleFunc("/api/packs", GetPacks)
	http.Handle("/", r)
	fmt.Println("listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello from backend!")
}

func GetPacks(w http.ResponseWriter, req *http.Request) {
	data := []pack{
		{
			Id:          0,
			Name:        "Bar Pack One",
			RoleCount:   3,
			StringCount: 27,
		},
		{
			Id:          1,
			Name:        "Foo Pack Two",
			RoleCount:   4,
			StringCount: 32,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
