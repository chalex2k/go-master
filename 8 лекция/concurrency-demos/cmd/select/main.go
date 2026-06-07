package main

import (
	"fmt"

	"lesson8/concurrency-demos/demos"
)

func main() {
	if err := demos.SelectWithContextDemo(); err != nil {
		fmt.Println(err)
	}
}
