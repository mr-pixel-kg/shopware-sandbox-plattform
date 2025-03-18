package main

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/config"
	"log"
	"net/http"
	"strconv"
)

// @title mpXsandbox API
// @version 1.0.0
// @description Management API for Docker Sandbox Enviroment
// @license.name MIT
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.basic BasicAuth
// @schemes http https
func main() {
	e := echo.New()

	// Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Database
	database.ConnectDB(cfg.Database)

	// Middleware
	e.Use(middleware.Logger())  // Loggt Anfragen
	e.Use(middleware.Recover()) // Fängt Panics ab und gibt 500 zurück
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.Server.AllowedOrigins,
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "Access-Control-Allow-Origin"},
	}))

	// Register routes
	api.RegisterRoutes(e, cfg)

	// Start server
	port := cfg.Server.Port
	log.Printf("Starting server on http://localhost:%d", port)
	if err := e.Start(":" + strconv.Itoa(port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Could not start server: %v", err)
	}
}
