package main

import (
	"log"

	"tourism-backend/config"
	"tourism-backend/internal/app"
)

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
