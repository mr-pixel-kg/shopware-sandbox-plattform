package api

import (
	"github.com/labstack/echo/v4"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler/images"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler/sandboxes"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/handler/system"
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
	guardService := services.NewGuardService(sessionRepository, config.Guard)
	auditLogService := services.NewAuditLogService(auditLogRepository)

	dockerService, err := services.NewDockerService()
	if err != nil {
		e.Logger.Fatalf("Failed to create Docker service: %v", err)
	}

	imageService, err := services.NewImageService(dockerService, imageRepository)
	if err != nil {
		e.Logger.Fatalf("Failed to create Image service: %v", err)
	}

	sandboxService, err := services.NewSandboxService(dockerService, imageService, guardService, sandboxRepository, *config)
	if err != nil {
		e.Logger.Fatalf("Failed to create Sandbox service: %v", err)
	}

	// Init handlers
	imageHandler := images.NewImageHandler(dockerService, imageService, auditLogService)
	sandboxHandler := sandboxes.NewSandboxHandler(sandboxService, auditLogService, guardService)
	systemHandler := system.NewSystemHandler(auditLogService)

	// Init middlewares
	//authMiddleware := middleware.OptionalAuthMiddleware(config.Auth)
	authRequiredMiddleware := middleware.AuthRequiredMiddleware(config.Auth)

	// Add api handlers
	api := e.Group("/api")

	api.Use(middleware.AuthMiddleware(config.Auth))

	api.GET("/health", handler.HealthCheckHandler)
	api.GET("/auth", handler.AuthCheckHandler)
	api.GET("/auditlog", systemHandler.SystemGetLastAuditLog, authRequiredMiddleware)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api.GET("/sandboxes", sandboxHandler.SandboxListHandler, authRequiredMiddleware)
	api.GET("/sandboxes/:id", sandboxHandler.SandboxDetailsHandler)
	api.POST("/sandboxes", sandboxHandler.SandboxCreateHandler)
	api.DELETE("/sandboxes/:id", sandboxHandler.SandboxDeleteHandler)
	api.POST("/sandboxes/:id/snapshot", sandboxHandler.SandboxSnapshotHandler, authRequiredMiddleware)

	api.GET("/images", imageHandler.ImageListHandler)
	api.GET("/images/:id", imageHandler.ImageDetailsHandler)
	api.POST("/images", imageHandler.PullImageHandler, authRequiredMiddleware)
	api.DELETE("/images/:id", imageHandler.ImageDeleteHandler, authRequiredMiddleware)
}
