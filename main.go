package main

import (
	"log"
)

func main() {
	config := DefaultConfig()
	client := NewClient(config)

	if err := client.Run(); err != nil {
		log.Fatalf("Client error: %v", err)
	}
}
