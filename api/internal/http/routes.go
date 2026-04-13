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
	"github.com/go-fuego/fuego/param"
	"github.com/mr-pixel-kg/shopshredder/api/internal/config"
	"github.com/mr-pixel-kg/shopshredder/api/internal/docker"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/handlers"
	mw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/registry"
	"github.com/mr-pixel-kg/shopshredder/api/internal/repositories"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
	"github.com/mr-pixel-kg/shopshredder/api/internal/sshproxy"
	"gorm.io/gorm"
)

type Server struct {
	fuego *fuego.Server
	cfg   config.Config
}

type runtimeServices struct {
	auth           *services.AuthService
	audit          *services.AuditService
	user           *services.UserService
	image          *services.ImageService
	sandbox        *services.SandboxService
	sandboxHealth  *services.SandboxHealthService
	terminal       *services.TerminalService
	registrySearch *services.RegistrySearchService
	log            *services.LogService
	resolver       *registry.Resolver
	userRepo       *repositories.UserRepository
	dockerSDK      *client.Client
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

	auth := handlers.AuthHandler{Auth: runtime.auth, Audit: runtime.audit}
	image := handlers.ImageHandler{Images: runtime.image, Audit: runtime.audit, Resolver: runtime.resolver}
	sandbox := handlers.SandboxHandler{Sandboxes: runtime.sandbox, Health: runtime.sandboxHealth, Auth: runtime.auth}
	audit := handlers.AuditHandler{Audit: runtime.audit}
	user := handlers.UserHandler{Users: runtime.user, Audit: runtime.audit}
	whitelist := handlers.WhitelistHandler{Users: runtime.user, Audit: runtime.audit}
	registrySearch := handlers.RegistrySearchHandler{Search: runtime.registrySearch}
	terminal := handlers.TerminalHandler{Terminals: runtime.terminal, Auth: runtime.auth, AllowedOrigins: cfg.Server.AllowedOrigins}
	log := handlers.LogHandler{Logs: runtime.log}

	bearerAuth := openapi3.SecurityRequirement{"bearerAuth": {}}

	public := fuego.Group(s, "/api")

	fuego.Post(public, "/auth/register", auth.Register,
		option.Summary("Register a new user"),
		option.Description("Create a new user account with email and password"),
		option.Tags("Auth"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Post(public, "/auth/login", auth.Login,
		option.Summary("Log in"),
		option.Description("Authenticate with email and password, receive a JWT token"),
		option.Tags("Auth"),
	)
	fuego.GetStd(public, "/images/{id}/progress", image.Progress,
		option.Summary("Stream image progress"),
		option.Description("SSE endpoint streaming progress events for image operations"),
		option.Tags("Images"),
	)
	fuego.Get(public, "/registry", image.RegistryLookup,
		option.Summary("Lookup registry metadata"),
		option.Description("Return registry-defined metadata for an image by name or ID"),
		option.Tags("Images"),
		option.Query("name", "Image name (e.g. dockware/dev:6.6.9.0)"),
		option.Query("id", "Image ID"),
	)
	fuego.GetStd(public, "/sandboxes/{id}/health", sandbox.HealthSSE,
		option.Summary("Stream sandbox health"),
		option.Description("SSE endpoint streaming sandbox readiness for active subscribers"),
		option.Tags("Sandboxes"),
		option.Query("access_token", "Bearer token fallback for EventSource"),
	)
	fuego.GetStd(public, "/sandboxes/{id}/stream", sandbox.StreamSSE,
		option.Summary("Stream sandbox state"),
		option.Description("SSE endpoint streaming real-time state updates for a single sandbox"),
		option.Tags("Sandboxes"),
		option.Query("access_token", "Bearer token fallback for EventSource"),
	)
	fuego.GetStd(public, "/sandboxes/{id}/terminal", terminal.Connect,
		option.Summary("Open interactive terminal session"),
		option.Description("Interactive shell (docker exec) into the sandbox container via WebSocket"),
		option.Tags("Sandboxes"),
		option.Query("access_token", "Bearer token"),
		option.QueryInt("cols", "Initial terminal columns (default 80)"),
		option.QueryInt("rows", "Initial terminal rows (default 24)"),
	)

	flex := fuego.Group(s, "/api",
		option.Middleware(mw.OptionalAuth(runtime.auth)),
	)

	fuego.Get(flex, "/sandboxes", sandbox.List,
		option.Summary("List sandboxes"),
		option.Description("Defaults to own sandboxes. Guests are identified by client_id cookie. Admins can pass ?scope=all to see all sandboxes."),
		option.Tags("Sandboxes"),
		option.Query("scope", "Admin only: 'all' to list all sandboxes across users"),
		option.QueryInt("limit", "Max entries per page (1-500, default 50)"),
		option.QueryInt("offset", "Offset for pagination (default 0)"),
	)
	fuego.Post(flex, "/sandboxes", sandbox.Create,
		option.Summary("Create a sandbox"),
		option.Description("Spin up a new sandbox. Guests are identified by a server-managed client_id cookie. Authenticated users get ownership via their user ID."),
		option.Tags("Sandboxes"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Delete(flex, "/sandboxes/{id}", sandbox.Delete,
		option.Summary("Delete a sandbox"),
		option.Description("Stop and remove a sandbox. Guests are identified by a server-managed client_id cookie. Authenticated users are checked by user ID or client_id cookie."),
		option.Tags("Sandboxes"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.Get(flex, "/images", image.ListImages,
		option.Summary("List images"),
		option.Description("Returns public images for guests or ?visibility=public. Authenticated users see all images."),
		option.Tags("Images"),
		option.Query("visibility", "Filter: 'public' for public images only"),
		option.QueryInt("limit", "Max entries per page (1-500, default 50)"),
		option.QueryInt("offset", "Offset for pagination (default 0)"),
	)

	authed := fuego.Group(s, "/api",
		option.Middleware(mw.Auth(runtime.auth)),
		option.Security(bearerAuth),
	)

	fuego.Post(authed, "/auth/logout", auth.Logout,
		option.Summary("Log out"),
		option.Description("Invalidate the current session token"),
		option.Tags("Auth"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.Get(authed, "/auth/me", auth.Me,
		option.Summary("Get current user"),
		option.Description("Return the authenticated user's profile"),
		option.Tags("Auth"),
	)
	fuego.Get(authed, "/sandboxes/{id}", sandbox.Get,
		option.Summary("Get sandbox by ID"),
		option.Description("Returns a single sandbox by its UUID"),
		option.Tags("Sandboxes"),
	)
	fuego.Patch(authed, "/sandboxes/{id}", sandbox.Update,
		option.Summary("Update sandbox"),
		option.Description("Update display name and/or extend TTL of a sandbox owned by the authenticated user"),
		option.Tags("Sandboxes"),
	)
	fuego.Post(authed, "/sandboxes/{id}/snapshots", sandbox.Snapshot,
		option.Summary("Create a snapshot image from a sandbox"),
		option.Description("Commit the current state of a running sandbox as a new Docker image"),
		option.Tags("Sandboxes"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Get(authed, "/images/pending", image.ListPending,
		option.Summary("List pending image operations"),
		option.Description("Returns all images with ongoing operations with optional progress percentage"),
		option.Tags("Images"),
	)
	fuego.Post(authed, "/images", image.Create,
		option.Summary("Create an image"),
		option.Description("Register a new Docker image. If not available locally, a background pull is started."),
		option.Tags("Images"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Patch(authed, "/images/{id}", image.Update,
		option.Summary("Update an image"),
		option.Description("Update image metadata and visibility"),
		option.Tags("Images"),
	)
	fuego.Delete(authed, "/images/{id}", image.Delete,
		option.Summary("Delete an image"),
		option.Description("Remove a Docker image registration"),
		option.Tags("Images"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.PostStd(authed, "/images/{id}/thumbnail", image.UploadThumbnail,
		option.Summary("Upload an image thumbnail"),
		option.Description("Upload or replace the thumbnail for an image. Send as multipart/form-data with field name 'thumbnail'."),
		option.Tags("Images"),
		option.RequestContentType("multipart/form-data"),
	)
	fuego.DeleteStd(authed, "/images/{id}/thumbnail", image.DeleteThumbnail,
		option.Summary("Delete an image thumbnail"),
		option.Description("Remove the thumbnail associated with an image"),
		option.Tags("Images"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.Get(authed, "/sandboxes/{id}/logs", log.ListSources,
		option.Summary("List available log sources"),
		option.Description("Returns all configured log sources for this sandbox's image"),
		option.Tags("Sandboxes"),
	)
	fuego.GetStd(authed, "/sandboxes/{id}/logs/{key}", log.StreamLog,
		option.Summary("Stream log output"),
		option.Description("SSE endpoint streaming live log output for a specific log source"),
		option.Tags("Sandboxes"),
	)
	fuego.Get(authed, "/registry/images/search", registrySearch.SearchImages,
		option.Summary("Search Docker Hub images"),
		option.Description("Returns matching image repositories from Docker Hub for autocomplete"),
		option.Tags("Registry"),
		option.Query("q", "Search query, minimum 2 characters (e.g. dockware/sh)", param.Required()),
	)
	fuego.Get(authed, "/registry/tags", registrySearch.SearchTags,
		option.Summary("Search Docker Hub tags"),
		option.Description("Returns matching tags for a Docker Hub image for autocomplete"),
		option.Tags("Registry"),
		option.Query("image", "Full image name (e.g. dockware/shopware)", param.Required()),
		option.Query("q", "Tag prefix filter (e.g. 6.7)"),
	)

	admin := fuego.Group(s, "/api",
		option.Middleware(mw.Auth(runtime.auth)),
		option.Middleware(mw.RequireAdmin()),
		option.Security(bearerAuth),
	)

	fuego.Get(admin, "/users", user.List,
		option.Summary("List users"),
		option.Description("Return all users, including pending invited users"),
		option.Tags("Users"),
		option.QueryInt("limit", "Max entries per page (1-500, default 50)"),
		option.QueryInt("offset", "Offset for pagination (default 0)"),
	)
	fuego.Get(admin, "/users/{id}", user.Get,
		option.Summary("Get user"),
		option.Description("Return a single user by ID"),
		option.Tags("Users"),
	)
	fuego.Post(admin, "/users", user.Create,
		option.Summary("Create user"),
		option.Description("Create an active user or invite a pending user when no password is provided"),
		option.Tags("Users"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Patch(admin, "/users/{id}", user.Update,
		option.Summary("Update user"),
		option.Description("Update a user's email, role, and optionally password"),
		option.Tags("Users"),
	)
	fuego.Delete(admin, "/users/{id}", user.Delete,
		option.Summary("Delete user"),
		option.Description("Delete a user by ID"),
		option.Tags("Users"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.Get(admin, "/whitelist", whitelist.List,
		option.Summary("List whitelisted emails"),
		option.Description("Return all pending (whitelisted but not yet registered) users"),
		option.Tags("Whitelist"),
	)
	fuego.Post(admin, "/whitelist", whitelist.Add,
		option.Summary("Add email to whitelist"),
		option.Description("Create a pending user row so the email can register in whitelist mode"),
		option.Tags("Whitelist"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Delete(admin, "/whitelist/{id}", whitelist.Remove,
		option.Summary("Remove email from whitelist"),
		option.Description("Delete a pending user row (only works for users that have not yet registered)"),
		option.Tags("Whitelist"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.Get(admin, "/audit-logs", audit.List,
		option.Summary("List audit logs"),
		option.Description("Returns recent audit log entries with pagination and filtering"),
		option.Tags("AuditLogs"),
		option.QueryInt("limit", "Max entries (1-500, default 50)"),
		option.QueryInt("offset", "Offset for pagination"),
		option.Query("userId", "Filter by user ID"),
		option.Query("action", "Filter by action"),
		option.Query("resourceType", "Filter by resource type"),
		option.Query("resourceId", "Filter by resource ID"),
		option.Query("clientId", "Filter by client ID"),
		option.Query("from", "Filter from timestamp (inclusive, RFC3339)"),
		option.Query("to", "Filter to timestamp (inclusive, RFC3339)"),
	)
	fuego.Get(admin, "/audit-logs/facets", audit.Facets,
		option.Summary("List audit log facets"),
		option.Description("Returns available audit filter values for the current query window"),
		option.Tags("AuditLogs"),
		option.Query("action", "Filter by action"),
		option.Query("resourceType", "Filter by resource type"),
		option.Query("resourceId", "Filter by resource ID"),
		option.Query("clientId", "Filter by client ID"),
		option.Query("from", "Filter from timestamp (inclusive, RFC3339)"),
		option.Query("to", "Filter to timestamp (inclusive, RFC3339)"),
	)
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
	sandboxHealthService := services.NewSandboxHealthService(sandboxRepo, imageRepo, resolver, executor)
	terminalService := services.NewTerminalService(cfg.Terminal, dockerClient, sandboxRepo)
	logService := services.NewLogService(dockerClient, sandboxRepo, imageRepo, resolver)

	registrySearchService := services.NewRegistrySearchService()

	return &runtimeServices{
		auth:           authService,
		audit:          auditService,
		user:           userService,
		image:          imageService,
		sandbox:        sandboxService,
		sandboxHealth:  sandboxHealthService,
		terminal:       terminalService,
		registrySearch: registrySearchService,
		log:            logService,
		resolver:       resolver,
		userRepo:       userRepo,
		dockerSDK:      sdkClient,
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
