package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/docker/docker/client"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/handlers"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/sshproxy"
	"gorm.io/gorm"
)

type Server struct {
	fuego *fuego.Server
	cfg   config.Config
}

type runtimeServices struct {
	auth          *services.AuthService
	audit         *services.AuditService
	user          *services.UserService
	image         *services.ImageService
	sandbox       *services.SandboxService
	sandboxHealth *services.SandboxHealthService
	terminal      *services.TerminalService
	log           *services.LogService
	resolver      *registry.Resolver
	userRepo      *repositories.UserRepository
	dockerSDK     *client.Client
}

func NewServer(cfg config.Config, db *gorm.DB) (*Server, error) {
	runtime, err := buildRuntimeServices(cfg, db)
	if err != nil {
		return nil, err
	}

	s := fuego.NewServer(
		fuego.WithAddr("0.0.0.0:"+strconv.Itoa(cfg.Server.Port)),
		fuego.WithoutAutoGroupTags(),
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				DisableLocalSave: true,
				PrettyFormatJSON: true,
				MiddlewareConfig: fuego.MiddlewareConfig{
					DisableMiddlewareSection: true,
				},
				Info: &openapi3.Info{
					Title:       "Shopware Sandbox Platform API",
					Description: "REST API for managing Shopware sandbox environments.",
					Version:     "1.0.0",
				},
			}),
		),
		fuego.WithSecurity(openapi3.SecuritySchemes{
			"bearerAuth": &openapi3.SecuritySchemeRef{
				Value: openapi3.NewSecurityScheme().
					WithType("http").
					WithScheme("bearer"),
			},
		}),
		fuego.WithGlobalMiddlewares(mw.CORS(cfg.Server.AllowedOrigins)),
	)
	s.WriteTimeout = 0

	startBackgroundJobs(cfg, runtime)
	registerRoutes(s, cfg, runtime)

	for _, pathItem := range s.OpenAPI.Description().Paths.Map() {
		for _, op := range pathItem.Operations() {
			filtered := make(openapi3.Parameters, 0, len(op.Parameters))
			for _, p := range op.Parameters {
				if p.Value != nil && p.Value.In == "header" && p.Value.Name == "Accept" {
					continue
				}
				filtered = append(filtered, p)
			}
			op.Parameters = filtered
		}
	}

	slog.Debug("http routes registered",
		"component", "http",
		"thumbnail_dir", cfg.Storage.ThumbnailDir,
	)

	return &Server{fuego: s, cfg: cfg}, nil
}

func registerRoutes(s *fuego.Server, cfg config.Config, runtime *runtimeServices) {
	fuego.Use(s, mw.RequestLogger())
	fuego.Use(s, mw.EnsureClientID())

	fuego.Get(s, "/health", healthCheck,
		option.Summary("Health check"),
		option.Tags("System"),
	)

	s.Mux.Handle(services.ThumbnailPublicBasePath+"/",
		http.StripPrefix(services.ThumbnailPublicBasePath, http.FileServer(http.Dir(cfg.Storage.ThumbnailDir))),
	)

	authHandler := handlers.AuthHandler{Auth: runtime.auth, Audit: runtime.audit}
	imageHandler := handlers.ImageHandler{Images: runtime.image, Audit: runtime.audit, Resolver: runtime.resolver}
	sandboxHandler := handlers.SandboxHandler{Sandboxes: runtime.sandbox, Health: runtime.sandboxHealth, Auth: runtime.auth}
	auditHandler := handlers.AuditHandler{Audit: runtime.audit}
	userHandler := handlers.UserHandler{Users: runtime.user, Audit: runtime.audit}
	whitelistHandler := handlers.WhitelistHandler{Users: runtime.user, Audit: runtime.audit}
	terminalHandler := handlers.TerminalHandler{Terminals: runtime.terminal, Auth: runtime.auth, AllowedOrigins: cfg.Server.AllowedOrigins}
	logHandler := handlers.LogHandler{Logs: runtime.log}

	public := fuego.Group(s, "/api")
	authHandler.MountPublicRoutes(public)
	imageHandler.MountPublicRoutes(public)
	sandboxHandler.MountPublicRoutes(public)
	terminalHandler.MountPublicRoutes(public)

	bearerAuth := openapi3.SecurityRequirement{"bearerAuth": {}}

	authed := fuego.Group(s, "/api",
		option.Middleware(mw.Auth(runtime.auth)),
		option.Security(bearerAuth),
	)
	authHandler.MountAuthedRoutes(authed)
	imageHandler.MountAuthedRoutes(authed)
	sandboxHandler.MountAuthedRoutes(authed)
	logHandler.MountAuthedRoutes(authed)

	admin := fuego.Group(s, "/api",
		option.Middleware(mw.Auth(runtime.auth)),
		option.Middleware(mw.RequireAdmin()),
		option.Security(bearerAuth),
	)
	userHandler.MountRoutes(admin)
	whitelistHandler.MountRoutes(admin)
	auditHandler.MountRoutes(admin)
}

func healthCheck(_ fuego.ContextNoBody) (dto.HealthResponse, error) {
	return dto.HealthResponse{Status: "ok"}, nil
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
	imageService := services.NewImageService(imageRepo, sandboxRepo, dockerClient, pullTracker, cfg.Server.BaseURL, cfg.Storage.ThumbnailDir, resolver)
	sandboxService := services.NewSandboxService(cfg.Sandbox, cfg.Docker, cfg.Guard, cfg.SSH, sandboxRepo, imageRepo, imageService, eventRepo, auditService, dockerClient, resolver, executor)
	sandboxHealthService := services.NewSandboxHealthService(sandboxRepo, imageRepo, resolver)
	terminalService := services.NewTerminalService(cfg.Terminal, dockerClient, sandboxRepo)
	logService := services.NewLogService(dockerClient, sandboxRepo, imageRepo, resolver)

	return &runtimeServices{
		auth:          authService,
		audit:         auditService,
		user:          userService,
		image:         imageService,
		sandbox:       sandboxService,
		sandboxHealth: sandboxHealthService,
		terminal:      terminalService,
		log:           logService,
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

func (s *Server) Start() error {
	slog.Info("starting http server", "port", s.cfg.Server.Port)
	return s.fuego.Run()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.fuego.Shutdown(ctx)
}
