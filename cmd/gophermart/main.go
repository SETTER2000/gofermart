package main

import (
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/app"
	"log"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
