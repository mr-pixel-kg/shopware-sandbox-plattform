// @title           Shopshredder Sandbox API
// @version         1.0.0
// @description     API for managing public demo sandboxes and internal employee sandboxes.
//
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Paste the JWT token. "Bearer " is optional.

//go:generate go tool swag init -g main.go -d .,../../internal/http/handlers,../../internal/http/dto,../../internal/models,../../internal/http -o ../../docs --parseInternal --packagePrefix github.com/manuel/shopware-testenv-platform/api

package main

import (
	"log"
	"log/slog"
	"os"

	_ "github.com/manuel/shopware-testenv-platform/api/docs"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/database"
	httpserver "github.com/manuel/shopware-testenv-platform/api/internal/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Load runtime configuration before touching external dependencies.
	cfg := config.MustLoad()
	slog.Info("configuration loaded",
		"port", cfg.Server.Port,
		"allowed_origins", cfg.Server.AllowedOrigins,
		"sandbox_default_ttl", cfg.Sandbox.DefaultTTL.String(),
		"sandbox_cleanup_interval", cfg.Sandbox.CleanupInterval.String(),
		"thumbnail_dir", cfg.Storage.ThumbnailDir,
	)

	// Open the shared database connection used by repositories and services.
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	slog.Info("database connection established", "host", cfg.Database.Host, "database", cfg.Database.Name)

	// Wire the HTTP server with all repositories, services and middleware.
	server, err := httpserver.NewServer(cfg, db)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}
	slog.Info("http server initialized", "base_url", cfg.Server.BaseURL, "port", cfg.Server.Port)

	// Start serving the API only after the full dependency graph is ready.
	if err := server.Start(); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
