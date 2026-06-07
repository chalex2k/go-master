package main

import (
	"fmt"

	"lesson8/concurrency-demos/demos"
)

func main() {
	if err := demos.MutexDemo(); err != nil {
		fmt.Println(err)
	}
}
