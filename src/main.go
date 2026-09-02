package main

import (
	"log"

	"github.com/CaiqueGOliveira/TelemetryGo/src/bootstrap"
)

func main() {
	router := bootstrap.AppBootstrap()

	log.Println("Server running at port 8080...")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Error on server initializing: %v", err)
	}
}
