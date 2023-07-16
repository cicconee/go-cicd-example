package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cicconee/go-cicd-example/pkg"
)

func main() {
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

	err := http.ListenAndServe(":8000", mux)
	log.Println("ERR:", err)
}

func Increment(i int) int {
	return i + 1
}
