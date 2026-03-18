package http

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/handlers"
	authmw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"gorm.io/gorm"
)

type Server struct {
	echo *echo.Echo
	cfg  config.Config
}

func NewServer(cfg config.Config, db *gorm.DB) (*Server, error) {
	e := echo.New()
	e.HideBanner = true
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(logging.EchoRequestLogger())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Repositories stay close to persistence concerns, while services own the
	// business rules built on top of them.
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	imageRepo := repositories.NewImageRepository(db)
	sandboxRepo := repositories.NewSandboxRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	eventRepo := repositories.NewSandboxEventRepository(db)

	passwordService := services.NewPasswordService()
	tokenService := services.NewTokenService(cfg.Auth)
	auditService := services.NewAuditService(auditRepo)
	authService := services.NewAuthService(userRepo, sessionRepo, passwordService, tokenService)
	guestService := services.NewGuestSessionService(sessionRepo, tokenService)
	dockerClient, err := docker.NewClient(cfg.Sandbox, cfg.Docker)
	if err != nil {
		return nil, err
	}
	imageService := services.NewImageService(imageRepo, dockerClient)
	sandboxService := services.NewSandboxService(cfg.Sandbox, cfg.Guard, sandboxRepo, imageRepo, imageService, eventRepo, auditService, dockerClient)

	// Sandbox expiration is handled inside the same process on purpose to keep
	// deployment simple for the single-service architecture.
	sandboxService.StartCleanupLoop(context.Background())

	authHandler := handlers.NewAuthHandler(authService, auditService)
	imageHandler := handlers.NewImageHandler(imageService, auditService)
	sandboxHandler := handlers.NewSandboxHandler(sandboxService)
	auditHandler := handlers.NewAuditHandler(auditService)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api")
	public := api.Group("/public")
	auth := api.Group("/auth")
	private := api.Group("")
	private.Use(authmw.Auth(authService))
	// Public routes create or refresh the guest cookie automatically.
	public.Use(authmw.EnsureGuestSession(guestService, cfg.Auth.GuestCookieName))

	public.GET("/images", imageHandler.ListPublic)
	public.GET("/sandboxes", sandboxHandler.ListGuest)
	public.POST("/demos", sandboxHandler.CreatePublicDemo)
	public.DELETE("/sandboxes/:id", sandboxHandler.DeleteGuest)

	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	private.POST("/auth/logout", authHandler.Logout)
	private.GET("/me", authHandler.Me)
	private.GET("/me/sandboxes", sandboxHandler.ListMine)
	private.GET("/images", imageHandler.ListAll)
	private.POST("/images", imageHandler.Create)
	private.DELETE("/images/:id", imageHandler.Delete)
	private.GET("/sandboxes", sandboxHandler.List)
	private.GET("/sandboxes/:id", sandboxHandler.Get)
	private.POST("/sandboxes", sandboxHandler.CreatePrivateSandbox)
	private.DELETE("/sandboxes/:id", sandboxHandler.Delete)
	private.POST("/sandboxes/:id/snapshot", sandboxHandler.Snapshot)
	private.GET("/audit-logs", auditHandler.List)

	e.File("/swagger/openapi.yaml", "docs/openapi.yaml")
	e.GET("/swagger", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({ url: '/swagger/openapi.yaml', dom_id: '#swagger-ui' });
  </script>
</body>
</html>`)
	})

	slog.Info("http routes registered",
		"public_routes", 4,
		"auth_routes", 2,
		"private_routes", 9,
		"guest_cookie_name", cfg.Auth.GuestCookieName,
	)

	return &Server{echo: e, cfg: cfg}, nil
}

func (s *Server) Start() error {
	slog.Info("starting http server", "port", s.cfg.Server.Port)
	return s.echo.Start(":" + strconv.Itoa(s.cfg.Server.Port))
}
