package main

import (
	"log"
	"os"
	"time"

	"github.com/cicconee/go-cicd-example/pkg"
)

func main() {
	count := 0
	name := os.Getenv("NAME")

	for {
		log.Printf("Count HAHA: %d\n", count)
		log.Println("NAME:", name)
		log.Println("Works with its own ssh key")
		log.Println("Push while logged out")
		time.Sleep(time.Second)
		count = pkg.Increment(count)
	}
}

func Increment(i int) int {
	return i + 1
}
