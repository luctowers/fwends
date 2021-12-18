package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type pack struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	RoleCount   int    `json:"roleCount"`
	StringCount int    `json:"stringCount"`
}

func main() {
	db, err := OpenPostgres()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/api", Hello)
	r.Handle("/api/packs", GetPacks(db))
	r.Handle("/api/test", Test(db))
	http.Handle("/", r)

	log.Println("binding to port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func OpenPostgres() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=disable", // TODO: add intraservice tls
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_DB"),
	)
	return sql.Open("postgres", connStr)
}

func Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello from backend!")
}

func Test(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT table_name
			FROM information_schema.tables
	 		WHERE table_schema='public'	AND table_type='BASE TABLE'
		`)
		if err != nil {
			log.Fatal(err)
		}
		tableNames := []string{}
		for rows.Next() {
			var name string
			err := rows.Scan(&name)
			if err != nil {
				log.Fatal(err)
			} else {
				tableNames = append(tableNames, name)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tableNames)
	}
	return http.HandlerFunc(fn)
}

func GetPacks(db *sql.DB) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
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
	return http.HandlerFunc(fn)
}
