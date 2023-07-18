package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cicconee/go-cicd-example/pkg"
	_ "github.com/lib/pq"
)

func main() {

	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "test_app", "password", "0.0.0.0", "5432", "weather_app_db")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ERR: %v\n", err)
	}

	count := 1
	name := os.Getenv("MY_NAME")
	log.Println("MY_NAME:", name)

	mux := http.NewServeMux()
	mux.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"message": "success dude!"}`))
		log.Printf("Request %d finished", count)
		count = pkg.Increment(count)
	}))

	err = http.ListenAndServe(":8000", mux)
	log.Println("ERR:", err)
}

func Increment(i int) int {
	return i + 1
}
