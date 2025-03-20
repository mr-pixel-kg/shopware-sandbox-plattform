package api

import (
	"github.com/labstack/echo/v4"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler/images"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler/sandboxes"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/config"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/repository"
	_ "github.com/mr-pixel-kg/shopware-sandbox-plattform/docs"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/middleware"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func RegisterRoutes(e *echo.Echo, config *config.Config) {

	// Init repositories
	imageRepository := repository.NewImageRepository(database.DB)
	sandboxRepository := repository.NewSandboxRepository(database.DB)
	auditLogRepository := repository.NewAuditLogRepository(database.DB)
	sessionRepository := repository.NewSessionRepository(database.DB)

	// Init services
	guardService := services.NewGuardService(sessionRepository)
	auditLogService := services.NewAuditLogService(auditLogRepository)

	dockerService, err := services.NewDockerService()
	if err != nil {
		e.Logger.Fatalf("Failed to create Docker service: %v", err)
	}

	imageService, err := services.NewImageService(imageRepository)
	if err != nil {
		e.Logger.Fatalf("Failed to create Image service: %v", err)
	}

	sandboxService, err := services.NewSandboxService(imageService, guardService, sandboxRepository)
	if err != nil {
		e.Logger.Fatalf("Failed to create Sandbox service: %v", err)
	}

	// Init handlers
	imageHandler := images.NewImageHandler(dockerService, imageService, auditLogService)
	sandboxHandler := sandboxes.NewSandboxHandler(sandboxService, auditLogService, guardService)

	// Init middlewares
	authMiddleware := middleware.OptionalAuthMiddleware(config.Auth)
	authRequiredMiddleware := middleware.AuthRequiredMiddleware(config.Auth)

	// Add api handlers
	api := e.Group("/api", authMiddleware)

	api.GET("/health", handler.HealthCheckHandler)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api.GET("/sandboxes", sandboxHandler.SandboxListHandler)
	api.GET("/sandboxes/:id", sandboxHandler.SandboxDetailsHandler)
	api.POST("/sandboxes", sandboxHandler.SandboxCreateHandler)
	api.DELETE("/sandboxes/:id", sandboxHandler.SandboxDeleteHandler, authRequiredMiddleware)

	api.GET("/images", imageHandler.ImageListHandler)
	api.GET("/images/:id", imageHandler.ImageDetailsHandler)
	api.POST("/images", imageHandler.PullImageHandler, authRequiredMiddleware)
	api.DELETE("/images/:id", imageHandler.ImageDeleteHandler, authRequiredMiddleware)
}
