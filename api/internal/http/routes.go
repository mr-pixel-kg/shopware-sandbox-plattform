package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/handlers"
	authmw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/sshproxy"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"
)

type Server struct {
	echo *echo.Echo
	cfg  config.Config
}

type runtimeServices struct {
	auth          *services.AuthService
	audit         *services.AuditService
	user          *services.UserService
	image         *services.ImageService
	sandbox       *services.SandboxService
	sandboxHealth *services.SandboxHealthService
	terminal      *services.TerminalService
	resolver      *registry.Resolver
	userRepo      *repositories.UserRepository
	dockerSDK     *client.Client
}

func NewServer(cfg config.Config, db *gorm.DB) (*Server, error) {
	e := newEcho(cfg)

	runtime, err := buildRuntimeServices(cfg, db)
	if err != nil {
		return nil, err
	}

	startBackgroundJobs(cfg, runtime)
	registerRoutes(e, cfg, runtime)

	slog.Debug("http routes registered",
		"component", "http",
		"thumbnail_dir", cfg.Storage.ThumbnailDir,
	)

	return &Server{echo: e, cfg: cfg}, nil
}

func newEcho(cfg config.Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = NewValidator()
	e.Use(authmw.EnsureClientID())
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(logging.EchoRequestLogger())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Client-Id"},
		AllowCredentials: true,
	}))
	return e
}

func buildRuntimeServices(cfg config.Config, db *gorm.DB) (*runtimeServices, error) {
	userRepo := repositories.NewUserRepository(db)
	imageRepo := repositories.NewImageRepository(db)
	sandboxRepo := repositories.NewSandboxRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	eventRepo := repositories.NewSandboxEventRepository(db)

	passwordService := services.NewPasswordService()
	tokenService := services.NewTokenService(cfg.Auth)
	auditService := services.NewAuditService(auditRepo)
	authService := services.NewAuthService(userRepo, passwordService, tokenService, cfg.Registration)
	userService := services.NewUserService(userRepo, passwordService)

	reg, err := registry.Load(cfg.RegistryPath)
	if err != nil {
		return nil, fmt.Errorf("load image registry: %w", err)
	}
	resolver, err := registry.NewResolver(reg)
	if err != nil {
		return nil, fmt.Errorf("compile image registry: %w", err)
	}

	// FIXME NewClientWithOpts ide says: Potential resource leak: ensure the resource is closed on all execution paths
	sdkClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	if _, err := sdkClient.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("docker daemon not reachable: %w", err)
	}

	dockerClient := docker.NewClient(sdkClient, cfg.Sandbox, cfg.Docker)
	executor := &registry.Executor{Client: sdkClient}
	pullTracker := docker.NewPullTracker()
	imageService := services.NewImageService(imageRepo, sandboxRepo, dockerClient, pullTracker, cfg.Server.BaseURL, cfg.Storage.ThumbnailDir)
	sandboxService := services.NewSandboxService(cfg.Sandbox, cfg.Docker, cfg.Guard, sandboxRepo, imageRepo, imageService, eventRepo, auditService, dockerClient, resolver, executor)
	sandboxHealthService := services.NewSandboxHealthService(sandboxRepo, imageRepo, resolver)
	terminalService := services.NewTerminalService(cfg.Terminal, dockerClient, sandboxRepo)

	return &runtimeServices{
		auth:          authService,
		audit:         auditService,
		user:          userService,
		image:         imageService,
		sandbox:       sandboxService,
		sandboxHealth: sandboxHealthService,
		terminal:      terminalService,
		resolver:      resolver,
		userRepo:      userRepo,
		dockerSDK:     sdkClient,
	}, nil
}

func startBackgroundJobs(cfg config.Config, runtime *runtimeServices) {
	runtime.sandbox.ReconcileOnStartup(context.Background())
	runtime.sandbox.StartCleanupLoop(context.Background())
	runtime.sandbox.StartDockerEventLoop(context.Background())
	runtime.sandboxHealth.StartMonitoringActive()
	runtime.image.ReconcileOnStartup(context.Background())

	if cfg.SSH.Enabled {
		srv := sshproxy.NewServer(
			fmt.Sprintf(":%d", cfg.SSH.Port),
			"",
			cfg.Sandbox.URLPrefix,
			runtime.dockerSDK,
			cfg.Docker.Network,
		)
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				slog.Error("SSH proxy server failed", "error", err)
			}
		}()
	}
}

func registerRoutes(e *echo.Echo, cfg config.Config, runtime *runtimeServices) {
	authHandler := handlers.NewAuthHandler(runtime.auth, runtime.audit)
	imageHandler := handlers.NewImageHandler(runtime.image, runtime.audit, runtime.resolver)
	sandboxHandler := handlers.NewSandboxHandler(
		runtime.sandbox,
		runtime.image,
		runtime.resolver,
		runtime.sandboxHealth,
		runtime.auth,
		cfg.SSH,
	)
	auditHandler := handlers.NewAuditHandler(runtime.audit)
	userHandler := handlers.NewUserHandler(runtime.user, runtime.audit)
	whitelistHandler := handlers.NewWhitelistHandler(runtime.userRepo, runtime.audit)
	terminalHandler := handlers.NewTerminalHandler(runtime.terminal, runtime.auth, cfg.Server.AllowedOrigins)

	e.GET("/health", healthCheck)
	registerAPIRoutes(e, runtime, authHandler, imageHandler, sandboxHandler, auditHandler, userHandler, whitelistHandler, terminalHandler)
	registerDocumentationRoutes(e, cfg)
}

func registerAPIRoutes(
	e *echo.Echo,
	runtime *runtimeServices,
	authHandler *handlers.AuthHandler,
	imageHandler *handlers.ImageHandler,
	sandboxHandler *handlers.SandboxHandler,
	auditHandler *handlers.AuditHandler,
	userHandler *handlers.UserHandler,
	whitelistHandler *handlers.WhitelistHandler,
	terminalHandler *handlers.TerminalHandler,
) {
	public := e.Group("/api")
	authed := e.Group("/api", authmw.Auth(runtime.auth))
	admin := e.Group("/api", authmw.Auth(runtime.auth), authmw.RequireAdmin())

	// public endpoints without auth
	public.POST("/auth/register", authHandler.Register)
	public.POST("/auth/login", authHandler.Login)
	public.GET("/images/:id/progress", imageHandler.Progress)
	public.GET("/registry", imageHandler.RegistryLookup)
	public.GET("/images/public", imageHandler.ListPublic)
	public.POST("/demos", sandboxHandler.CreateDemo)
	public.GET("/demos", sandboxHandler.ListDemos)
	public.DELETE("/demos/:id", sandboxHandler.DeleteDemo)
	public.GET("/sandboxes/:id/health", sandboxHandler.Health)
	public.GET("/sandboxes/:id/stream", sandboxHandler.Stream)
	public.GET("/sandboxes/:id/terminal", terminalHandler.Connect)

	// private endpoints with auth required
	authed.POST("/auth/logout", authHandler.Logout)
	authed.GET("/auth/me", authHandler.Me)

	authed.GET("/images", imageHandler.ListAll)
	authed.GET("/images/pending", imageHandler.ListPending)
	authed.POST("/images", imageHandler.Create)
	authed.PUT("/images/:id", imageHandler.Update)
	authed.DELETE("/images/:id", imageHandler.Delete)
	authed.POST("/images/:id/thumbnail", imageHandler.UploadThumbnail)
	authed.DELETE("/images/:id/thumbnail", imageHandler.DeleteThumbnail)

	authed.GET("/sandboxes", sandboxHandler.List)
	authed.GET("/sandboxes/:id", sandboxHandler.Get)
	authed.POST("/sandboxes", sandboxHandler.Create)
	authed.PATCH("/sandboxes/:id", sandboxHandler.Update)
	authed.DELETE("/sandboxes/:id", sandboxHandler.Delete)
	authed.POST("/sandboxes/:id/snapshots", sandboxHandler.Snapshot)

	// private endpoints with auth required + admin role
	admin.GET("/users", userHandler.List)
	admin.GET("/users/:id", userHandler.Get)
	admin.POST("/users", userHandler.Create)
	admin.PATCH("/users/:id", userHandler.Update)
	admin.DELETE("/users/:id", userHandler.Delete)
	admin.GET("/whitelist", whitelistHandler.List)
	admin.POST("/whitelist", whitelistHandler.Add)
	admin.DELETE("/whitelist/:id", whitelistHandler.Remove)
	admin.GET("/audit-logs", auditHandler.List)
	admin.GET("/audit-logs/facets", auditHandler.Facets)
}

func registerDocumentationRoutes(e *echo.Echo, cfg config.Config) {
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	e.Static(services.ThumbnailPublicBasePath, cfg.Storage.ThumbnailDir)
	e.GET("/docs/*", echoSwagger.WrapHandler)
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

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
