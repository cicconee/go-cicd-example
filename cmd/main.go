package main

import (
	"fmt"
	"time"
)

func main() {
	count := 0

	for {
		fmt.Printf("Count: %d\n", count)
		time.Sleep(time.Second)
		count++
	}
}
