// Package app configures and runs application.
package main

import (
	"log"
	"social-service/config"
	"social-service/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
