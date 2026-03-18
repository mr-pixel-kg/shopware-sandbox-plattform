package main

import (
	"log"

	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/database"
	httpserver "github.com/manuel/shopware-testenv-platform/api/internal/http"
)

func main() {
	// Load runtime configuration before touching external dependencies.
	cfg := config.MustLoad()

	// Open the shared database connection used by repositories and services.
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	// Wire the HTTP server with all repositories, services and middleware.
	server, err := httpserver.NewServer(cfg, db)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	// Start serving the API only after the full dependency graph is ready.
	if err := server.Start(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
