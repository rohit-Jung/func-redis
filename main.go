// Package main -
package main

import (
	"log"

	"github.com/rohit-Jung/func-redis/server"
)

func main() {
	err := server.RunAsyncServer()
	if err != nil {
		log.Fatal("Error running server", err)
	}
}
