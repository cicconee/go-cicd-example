package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("Hello,", os.Getenv("MY_NAME"))

	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "app", "password", "0.0.0.0", "5432", "app")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalln(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("DB ERR: %v\n", err)
	}

	handler := Handler{db: db}

	count := 1

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("RemoteAddr: %s\n", r.RemoteAddr)

		switch r.Method {
		case "GET":
			handler.Get(w, r)
		case "POST":
			handler.Post(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte(`{
				"message": "Method not implemented"
			}`))
		}

		log.Printf("Request %d finished\n", count)
		count = Increment(count)
	}))

	err = http.ListenAndServe(":8000", mux)
	log.Println("ERR:", err)
}

func Increment(i int) int {
	return i + 1
}

type Handler struct {
	db *sql.DB
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("Parsing integer: %v\n", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{
			"message": "Internal server error"
		}`))
		return
	}

	var name string
	var age int

	if err := h.db.QueryRowContext(r.Context(), "SELECT name, age FROM users WHERE id = $1", id).
		Scan(&name, &age); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("User (id=%d) not found\n", id)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			w.Write([]byte(`{
					"message": "User does not exist"
				}`))
			return
		}

		log.Printf("Querying user (id=%d): %v\n", id, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{
			"message": "Internal server error"
		}`))
		return
	}

	resp := fmt.Sprintf(`{
		"message": "User found",
		"user": {
			"id": %d,
			"name": "%s",
			"age": %d
		}
	}`, id, name, age)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(resp))
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	_, err := h.db.ExecContext(r.Context(), "INSERT INTO users(id, name, age) VALUES($1, $2, $3)",
		10,
		"YOUR NAME",
		29,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{
			"message": "Internal server error"
		}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{
		"message": "Row inserted!"
	}`))
}
