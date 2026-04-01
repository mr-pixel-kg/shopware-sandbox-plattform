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
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"
)

type Server struct {
	echo *echo.Echo
	cfg  config.Config
}

type runtimeServices struct {
	auth          *services.AuthService
	guest         *services.GuestSessionService
	audit         *services.AuditService
	user          *services.UserService
	image         *services.ImageService
	sandbox       *services.SandboxService
	sandboxHealth *services.SandboxHealthService
	terminal      *services.TerminalService
	resolver      *registry.Resolver
	userRepo      *repositories.UserRepository
}

func NewServer(cfg config.Config, db *gorm.DB) (*Server, error) {
	e := newEcho(cfg)

	runtime, err := buildRuntimeServices(cfg, db)
	if err != nil {
		return nil, err
	}

	startBackgroundJobs(runtime)
	registerRoutes(e, cfg, runtime)

	slog.Debug("http routes registered",
		"component", "http",
		"guest_cookie_name", cfg.Auth.GuestCookieName,
		"thumbnail_dir", cfg.Storage.ThumbnailDir,
	)

	return &Server{echo: e, cfg: cfg}, nil
}

func newEcho(cfg config.Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Validator = NewValidator()
	e.Use(echomw.Recover())
	e.Use(echomw.RequestID())
	e.Use(logging.EchoRequestLogger())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	return e
}

func buildRuntimeServices(cfg config.Config, db *gorm.DB) (*runtimeServices, error) {
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	imageRepo := repositories.NewImageRepository(db)
	sandboxRepo := repositories.NewSandboxRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)
	eventRepo := repositories.NewSandboxEventRepository(db)

	passwordService := services.NewPasswordService()
	tokenService := services.NewTokenService(cfg.Auth)
	auditService := services.NewAuditService(auditRepo)
	authService := services.NewAuthService(userRepo, sessionRepo, passwordService, tokenService, cfg.Registration)
	userService := services.NewUserService(userRepo, passwordService)
	guestService := services.NewGuestSessionService(sessionRepo, tokenService)

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
		guest:         guestService,
		audit:         auditService,
		user:          userService,
		image:         imageService,
		sandbox:       sandboxService,
		sandboxHealth: sandboxHealthService,
		terminal:      terminalService,
		resolver:      resolver,
		userRepo:      userRepo,
	}, nil
}

func startBackgroundJobs(runtime *runtimeServices) {
	runtime.sandbox.ReconcileOnStartup(context.Background())
	runtime.sandbox.StartCleanupLoop(context.Background())
	runtime.sandbox.StartDockerEventLoop(context.Background())
	runtime.sandboxHealth.StartMonitoringActive()
	runtime.image.ReconcileOnStartup(context.Background())
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
		runtime.guest,
		cfg.Auth.GuestCookieName,
	)
	auditHandler := handlers.NewAuditHandler(runtime.audit)
	userHandler := handlers.NewUserHandler(runtime.user, runtime.audit)
	whitelistHandler := handlers.NewWhitelistHandler(runtime.userRepo, runtime.audit)
	terminalHandler := handlers.NewTerminalHandler(runtime.terminal, runtime.auth, cfg.Server.AllowedOrigins)

	e.GET("/health", healthCheck)
	registerAPIRoutes(e, cfg, runtime, authHandler, imageHandler, sandboxHandler, auditHandler, userHandler, whitelistHandler, terminalHandler)
	registerDocumentationRoutes(e, cfg)
}

func registerAPIRoutes(
	e *echo.Echo,
	cfg config.Config,
	runtime *runtimeServices,
	authHandler *handlers.AuthHandler,
	imageHandler *handlers.ImageHandler,
	sandboxHandler *handlers.SandboxHandler,
	auditHandler *handlers.AuditHandler,
	userHandler *handlers.UserHandler,
	whitelistHandler *handlers.WhitelistHandler,
	terminalHandler *handlers.TerminalHandler,
) {
	api := e.Group("/api")
	public := api.Group("/public")
	auth := api.Group("/auth")
	private := api.Group("")
	private.Use(authmw.Auth(runtime.auth))
	public.Use(authmw.EnsureGuestSession(runtime.guest, cfg.Auth.GuestCookieName))

	api.GET("/images/:id/progress", imageHandler.Progress)
	api.GET("/registry/lookup", imageHandler.RegistryLookup)
	api.GET("/sandboxes/:id/health", sandboxHandler.Health)
	api.GET("/sandboxes/:id/terminal", terminalHandler.Connect)

	public.GET("/images", imageHandler.ListPublic)
	public.GET("/sandboxes", sandboxHandler.ListGuest)
	public.POST("/demos", sandboxHandler.CreatePublicDemo)
	public.DELETE("/sandboxes/:id", sandboxHandler.DeleteGuest)

	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	private.POST("/auth/logout", authHandler.Logout)
	private.GET("/me", authHandler.Me)
	private.GET("/me/sandboxes", sandboxHandler.ListMine)
	private.GET("/sandboxes/:id/stream", sandboxHandler.Stream)
	private.GET("/images", imageHandler.ListAll)
	private.GET("/images/pending", imageHandler.ListPending)
	private.POST("/images", imageHandler.Create)
	private.PUT("/images/:id", imageHandler.Update)
	private.DELETE("/images/:id", imageHandler.Delete)
	private.POST("/images/:id/thumbnail", imageHandler.UploadThumbnail)
	private.DELETE("/images/:id/thumbnail", imageHandler.DeleteThumbnail)
	private.GET("/sandboxes", sandboxHandler.List)
	private.GET("/sandboxes/:id", sandboxHandler.Get)
	private.POST("/sandboxes", sandboxHandler.CreatePrivateSandbox)
	private.DELETE("/sandboxes/:id", sandboxHandler.Delete)
	private.PATCH("/sandboxes/:id", sandboxHandler.Update)
	private.PATCH("/sandboxes/:id/ttl", sandboxHandler.ExtendTTL)
	private.POST("/sandboxes/:id/snapshot", sandboxHandler.Snapshot)
	private.GET("/audit-logs", auditHandler.List)

	admin := private.Group("/admin")
	admin.Use(authmw.RequireAdmin())
	admin.GET("/users", userHandler.List)
	admin.GET("/users/:id", userHandler.Get)
	admin.POST("/users", userHandler.Create)
	admin.PUT("/users/:id", userHandler.Update)
	admin.DELETE("/users/:id", userHandler.Delete)
	admin.GET("/whitelist", whitelistHandler.List)
	admin.POST("/whitelist", whitelistHandler.Add)
	admin.DELETE("/whitelist/:id", whitelistHandler.Remove)
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
