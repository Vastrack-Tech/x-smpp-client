package main

import (
	"log"

	"x-smpp-client/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(""); err != nil {
		log.Fatalf("bootstrap: %v", err)
	}
}
