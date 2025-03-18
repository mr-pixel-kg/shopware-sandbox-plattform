package api

import (
	"github.com/labstack/echo/v4"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database/repository"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler/images"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler/sandboxes"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/config"
	_ "github.com/mr-pixel-kg/shopware-sandbox-plattform/docs"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/middleware"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services"
	images2 "github.com/mr-pixel-kg/shopware-sandbox-plattform/services/images"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services/sandbox"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func RegisterRoutes(e *echo.Echo, config *config.Config) {

	// Init repositories
	imageRepository := repository.NewImageRepository(database.DB)
	sandboxRepository := repository.NewSandboxRepository(database.DB)

	// Init services
	dockerService, err := services.NewDockerService()
	if err != nil {
		e.Logger.Fatalf("Failed to create Docker service: %v", err)
	}

	imageService, err := images2.NewImageService(imageRepository)
	if err != nil {
		e.Logger.Fatalf("Failed to create Image service: %v", err)
	}

	sandboxService, err := sandbox.NewSandboxService(imageService, sandboxRepository)
	if err != nil {
		e.Logger.Fatalf("Failed to create Sandbox service: %v", err)
	}

	// Init handlers
	imageHandler := images.NewImageHandler(dockerService, imageService)
	sandboxHandler := sandboxes.NewSandboxHandler(sandboxService)

	// Init middlewares
	authMiddleware := middleware.AuthMiddleware(config.Auth)

	// Add api handlers
	api := e.Group("/api")

	api.GET("/health", handler.HealthCheckHandler)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api.GET("/sandboxes", sandboxHandler.SandboxListHandler)
	api.GET("/sandboxes/:id", sandboxHandler.SandboxDetailsHandler)
	api.POST("/sandboxes", sandboxHandler.SandboxCreateHandler)
	api.DELETE("/sandboxes/:id", sandboxHandler.SandboxDeleteHandler, authMiddleware)

	api.GET("/images", imageHandler.ImageListHandler)
	api.GET("/images/:id", imageHandler.ImageDetailsHandler)
	api.POST("/images", imageHandler.PullImageHandler, authMiddleware)
	api.DELETE("/images/:id", imageHandler.ImageDeleteHandler, authMiddleware)
}
