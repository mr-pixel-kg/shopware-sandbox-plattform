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

//go:generate go tool swag init -g main.go -d .,../../internal/http/handlers,../../internal/http/dto,../../internal/models,../../internal/http,../../internal/registry -o ../../docs --parseInternal --packagePrefix github.com/manuel/shopware-testenv-platform/api

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/manuel/shopware-testenv-platform/api/docs"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/database"
	httpserver "github.com/manuel/shopware-testenv-platform/api/internal/http"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
)

func main() {
	// Load runtime configuration before touching external dependencies.
	cfg := config.MustLoad()

	logging.Setup(cfg.Logging)

	slog.Info("configuration loaded",
		"port", cfg.Server.Port,
		"allowed_origins", cfg.Server.AllowedOrigins,
		"sandbox_default_ttl", cfg.Sandbox.DefaultTTL.String(),
		"sandbox_cleanup_interval", cfg.Sandbox.CleanupInterval.String(),
		"thumbnail_dir", cfg.Storage.ThumbnailDir,
		"log_level", cfg.Logging.Level,
		"log_format", cfg.Logging.Format,
	)

	// Open the shared database connection used by repositories and services.
	db, err := database.Connect(cfg.Database, logging.ParseLevel(cfg.Logging.Level))
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connection established", "host", cfg.Database.Host, "database", cfg.Database.Name)

	// Wire the HTTP server with all repositories, services and middleware.
	server, err := httpserver.NewServer(cfg, db)
	if err != nil {
		slog.Error("create server", "error", err)
		os.Exit(1)
	}
	slog.Info("http server initialized", "base_url", cfg.Server.BaseURL, "port", cfg.Server.Port)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	serverErrCh := make(chan error, 1)

	// Start serving the API only after the full dependency graph is ready.
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
	}()

	select {
	case err := <-serverErrCh:
		stop()
		slog.Error("http server error", "error", err)
		os.Exit(1)
	case <-ctx.Done():
		stop()
		slog.Info("shutdown signal received, draining connections")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		cancel()
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	cancel()
	slog.Info("server stopped gracefully")
}
