package main

import (
	"fmt"

	"lesson8/concurrency-demos/demos"
)

func main() {
	if err := demos.ErrGroupDemo(); err != nil {
		fmt.Println(err)
	}
}
