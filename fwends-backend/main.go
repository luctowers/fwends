package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
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

	rdb := OpenRedis()

	r := mux.NewRouter()
	r.HandleFunc("/api", Hello)
	r.Handle("/api/packs", GetPacks(db))
	r.Handle("/api/test1", Test1(db))
	r.Handle("/api/test2", Test2(rdb))
	http.Handle("/", r)

	log.Println("binding to port 80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func OpenPostgres() (*sql.DB, error) {
	// TODO: make postgres port configurable
	// TODO: add intraservice tls
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_DB"),
	)
	return sql.Open("postgres", connStr)
}

func OpenRedis() *redis.Client {
	// TODO: make redis port configurable
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(
			"%s:6379",
			os.Getenv("REDIS_HOST"),
		),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello from backend!")
}

func Test1(db *sql.DB) http.HandlerFunc {
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

func Test2(rdb *redis.Client) http.HandlerFunc {
	ctx := context.Background()
	fn := func(w http.ResponseWriter, r *http.Request) {
		result, err := rdb.Incr(ctx, "x").Result()
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
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
