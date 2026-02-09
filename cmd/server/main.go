package main

import (
	"log"

	"todo-app/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create application instance
	app, err := NewApplication(cfg)
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	// Start the server
	serverErrors := app.Start()

	// Wait for shutdown signal
	app.WaitForShutdown(serverErrors)

	// Gracefully shutdown
	if err := app.Shutdown(); err != nil {
		log.Fatal("Failed to shutdown gracefully:", err)
	}
}
