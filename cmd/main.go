package main

import (
	"fmt"
	"time"

	"github.com/cicconee/go-cicd-example/pkg"
)

func main() {
	count := 0

	for {
		fmt.Printf("Count: %d\n", count)
		time.Sleep(time.Second)
		count = pkg.Increment(count)
	}
}

func Increment(i int) int {
	return i + 1
}