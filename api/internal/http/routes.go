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
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/handlers"
	authmw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	echoSwagger "github.com/swaggo/echo-swagger"
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
	pullTracker := docker.NewPullTracker()
	imageService := services.NewImageService(imageRepo, sandboxRepo, dockerClient, pullTracker)
	sandboxService := services.NewSandboxService(cfg.Sandbox, cfg.Docker, cfg.Guard, sandboxRepo, imageRepo, imageService, eventRepo, auditService, dockerClient)

	// Sandbox expiration is handled inside the same process on purpose to keep
	// deployment simple for the single-service architecture.
	sandboxService.StartCleanupLoop(context.Background())

	authHandler := handlers.NewAuthHandler(authService, auditService)
	imageHandler := handlers.NewImageHandler(imageService, auditService)
	sandboxHandler := handlers.NewSandboxHandler(sandboxService)
	auditHandler := handlers.NewAuditHandler(auditService)

	e.GET("/health", healthCheck)

	api := e.Group("/api")
	public := api.Group("/public")
	auth := api.Group("/auth")
	private := api.Group("")
	private.Use(authmw.Auth(authService))
	// Public routes create or refresh the guest cookie automatically.
	public.Use(authmw.EnsureGuestSession(guestService, cfg.Auth.GuestCookieName))

	api.GET("/images/:id/progress", imageHandler.PullProgress)

	private.GET("/images/pulls", imageHandler.ListPulls)
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
	private.PATCH("/sandboxes/:id/ttl", sandboxHandler.ExtendTTL)
	private.POST("/sandboxes/:id/snapshot", sandboxHandler.Snapshot)
	private.GET("/audit-logs", auditHandler.List)

	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	e.GET("/docs/*", echoSwagger.WrapHandler)

    // TODO more dynamic public and private and auth routes instead of manually maintain these ints
	slog.Info("http routes registered",
		"public_routes", 4,
		"auth_routes", 2,
		"private_routes", 10,
		"guest_cookie_name", cfg.Auth.GuestCookieName,
	)

	return &Server{echo: e, cfg: cfg}, nil
}

// healthCheck godoc
// @Summary      Health check
// @Description  Returns service health status
// @Tags         System
// @Produce      json
// @Success      200 {object} dto.HealthResponse
// @Router       /health [get]
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}

func (s *Server) Start() error {
	slog.Info("starting http server", "port", s.cfg.Server.Port)
	return s.echo.Start(":" + strconv.Itoa(s.cfg.Server.Port))
}
