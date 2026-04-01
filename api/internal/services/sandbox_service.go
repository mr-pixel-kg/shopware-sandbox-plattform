package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrSandboxLimitReached = errors.New("sandbox limit reached")
	ErrSandboxNotFound     = errors.New("sandbox not found")
	ErrSandboxAccessDenied = errors.New("sandbox access denied")
)

type SandboxService struct {
	cfg       config.SandboxConfig
	dockerCfg config.DockerConfig
	guard     config.GuardConfig
	repo      *repositories.SandboxRepository
	imageRepo *repositories.ImageRepository
	images    *ImageService
	eventRepo *repositories.SandboxEventRepository
	audit     *AuditService
	docker    docker.Client
	resolver  *registry.Resolver
	executor  *registry.Executor
}

func NewSandboxService(
	cfg config.SandboxConfig,
	dockerCfg config.DockerConfig,
	guard config.GuardConfig,
	repo *repositories.SandboxRepository,
	imageRepo *repositories.ImageRepository,
	images *ImageService,
	eventRepo *repositories.SandboxEventRepository,
	audit *AuditService,
	dockerClient docker.Client,
	resolver *registry.Resolver,
	executor *registry.Executor,
) *SandboxService {
	return &SandboxService{
		cfg:       cfg,
		dockerCfg: dockerCfg,
		guard:     guard,
		repo:      repo,
		imageRepo: imageRepo,
		images:    images,
		eventRepo: eventRepo,
		audit:     audit,
		docker:    dockerClient,
		resolver:  resolver,
		executor:  executor,
	}
}

type CreateSandboxInput struct {
	ImageID        uuid.UUID
	UserID         *uuid.UUID
	GuestSessionID *uuid.UUID
	ClientIP       string
	TTLMinutes     *int
	DisplayName    *string
	Metadata       map[string]string
	AuditActor     AuditActor
}

type UpdateSandboxInput struct {
	SandboxID   uuid.UUID
	UserID      *uuid.UUID
	DisplayName *string
	ClientIP    string
	AuditActor  AuditActor
}

func (s *SandboxService) ListAll() ([]models.Sandbox, error) {
	return s.repo.ListAll()
}

func (s *SandboxService) ListActive() ([]models.Sandbox, error) {
	return s.repo.ListAllActive()
}

func (s *SandboxService) ListByUser(userID uuid.UUID) ([]models.Sandbox, error) {
	return s.repo.ListAllByUser(userID)
}

func (s *SandboxService) ListByGuestSession(sessionID uuid.UUID) ([]models.Sandbox, error) {
	return s.repo.ListAllByGuestSession(sessionID)
}

func (s *SandboxService) FindByID(id uuid.UUID) (*models.Sandbox, error) {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrSandboxNotFound
	}
	return sandbox, nil
}

func (s *SandboxService) UpdateSandbox(input UpdateSandboxInput) (*models.Sandbox, error) {
	sandbox, err := s.repo.FindByID(input.SandboxID)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	if input.UserID == nil || sandbox.OwnerID == nil || *sandbox.OwnerID != *input.UserID {
		return nil, ErrSandboxAccessDenied
	}

	if input.DisplayName != nil {
		sandbox.DisplayName = *input.DisplayName
	}

	if err := s.repo.Update(sandbox); err != nil {
		return nil, err
	}

	resourceType := auditcontracts.ResourceTypeSandbox
	_ = s.audit.Log(AuditLogInput{
		Actor:        input.AuditActor,
		Action:       auditcontracts.ActionSandboxUpdated,
		ResourceType: &resourceType,
		ResourceID:   &sandbox.ID,
		Details: map[string]any{
			"displayName": sandbox.DisplayName,
		},
	})

	return sandbox, nil
}

func (s *SandboxService) Create(ctx context.Context, input CreateSandboxInput) (*models.Sandbox, error) {
	if err := s.enforceLimits(input); err != nil {
		return nil, err
	}

	// Sandbox creation always starts from a registered image record so the
	// platform can show metadata and audit trails consistently.
	image, err := s.imageRepo.FindByID(input.ImageID)
	if err != nil {
		return nil, err
	}

	if image.Status == models.ImageStatusPulling {
		return nil, fmt.Errorf("image is still being pulled")
	}
	if image.Status == models.ImageStatusFailed {
		errMsg := ""
		if image.Error != nil {
			errMsg = *image.Error
		}
		return nil, fmt.Errorf("image pull failed: %s", errMsg)
	}
	if err := s.docker.EnsureImage(ctx, image.FullName()); err != nil {
		return nil, err
	}

	sandboxID := uuid.New()
	containerName := fmt.Sprintf("%s%s", s.cfg.URLPrefix, sandboxID.String())

	var expiresAt *time.Time
	var ttl time.Duration
	if input.TTLMinutes != nil && *input.TTLMinutes == 0 {
		ttl = 0
	} else {
		ttl = s.cfg.DefaultTTL
		if input.TTLMinutes != nil {
			ttl = time.Duration(*input.TTLMinutes) * time.Minute
		}
		if ttl > s.cfg.MaxTTL {
			ttl = s.cfg.MaxTTL
		}
		t := time.Now().UTC().Add(ttl)
		expiresAt = &t
	}

	// builds registry Configuration: env vars, labels, lifecycle hooks etc.
	var hostname string
	var hostPort int
	if s.dockerCfg.Mode == config.DockerModePort {
		hostPort, err = docker.FindFreePort()
		if err != nil {
			return nil, fmt.Errorf("find free port: %w", err)
		}
		hostname = fmt.Sprintf("localhost:%d", hostPort)
	} else {
		hostname = fmt.Sprintf("%s%s", containerName, s.cfg.HostSuffix)
	}

	expiresAtStr := ""
	if expiresAt != nil {
		expiresAtStr = expiresAt.Format(time.RFC3339)
	}
	tmplCtx := s.buildTemplateContext(
		image.FullName(), containerName, hostname,
		sandboxID.String(), ttl.String(), expiresAtStr,
		input.ClientIP, strconv.Itoa(hostPort), input.Metadata,
	)

	registryRef := image.RegistryName()
	resolved, err := s.resolver.Resolve(registryRef, tmplCtx)
	if err != nil {
		return nil, fmt.Errorf("resolve registry for %s: %w", image.FullName(), err)
	}

	internalPort := s.cfg.InternalPort
	if resolved.InternalPort > 0 {
		internalPort = resolved.InternalPort
	}

	// build labels: sandbox marker + resolved labels + traefik labels.
	labels := map[string]string{"sandbox_container": "true"}
	for k, v := range resolved.Labels {
		labels[k] = v
	}
	if s.dockerCfg.Mode == config.DockerModeTraefik {
		for k, v := range docker.BuildTraefikLabels(containerName, hostname, internalPort, s.dockerCfg) {
			labels[k] = v
		}
	}

	container, err := s.docker.CreateContainer(ctx, docker.ContainerCreateRequest{
		ImageName:     image.FullName(),
		ContainerName: containerName,
		Hostname:      hostname,
		Env:           resolved.Env,
		Labels:        labels,
		InternalPort:  internalPort,
	})
	if err != nil {
		return nil, err
	}

	if len(resolved.PostStart) > 0 {
		go s.executor.RunPostStart(context.Background(), container.ID, resolved.PostStart)
	}

	fieldsJSON, _ := json.Marshal(input.Metadata)
	displayName := ""
	if input.DisplayName != nil {
		displayName = *input.DisplayName
	}
	startingReason := "Container wird gestartet"
	sandbox := &models.Sandbox{
		ID:             sandboxID,
		ImageID:        image.ID,
		OwnerID:        input.UserID,
		GuestSessionID: input.GuestSessionID,
		DisplayName:    displayName,
		Status:         models.SandboxStatusStarting,
		StateReason:    &startingReason,
		ContainerID:    container.ID,
		ContainerName:  container.Name,
		URL:            container.URL,
		Port:           container.Port,
		ClientIP:       input.ClientIP,
		Metadata:       datatypes.JSON(fieldsJSON),
		ExpiresAt:      expiresAt,
	}

	if err := s.repo.Create(sandbox); err != nil {
		return nil, err
	}

	// Persisting a sandbox event gives the admin area a lightweight lifecycle log
	// without introducing a separate event bus.
	if err := s.addEvent(sandbox.ID, "created", map[string]any{
		"image": image.FullName(),
		"actor": sandboxActorType(sandbox),
	}); err != nil {
		return nil, err
	}

	resourceType := auditcontracts.ResourceTypeSandbox
	_ = s.audit.Log(AuditLogInput{
		Actor:        input.AuditActor,
		Action:       auditcontracts.ActionSandboxCreated,
		ResourceType: &resourceType,
		ResourceID:   &sandbox.ID,
		Details: map[string]any{
			"imageId": image.ID.String(),
			"actor":   sandboxActorType(sandbox),
		},
	})

	return sandbox, nil
}

func (s *SandboxService) ExtendTTL(id uuid.UUID, ttlMinutes *int, auditActor AuditActor) (*models.Sandbox, error) {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	if sandbox.Status != models.SandboxStatusRunning && sandbox.Status != models.SandboxStatusStarting {
		return nil, ErrSandboxNotFound
	}

	switch {
	case ttlMinutes == nil:
		return sandbox, nil
	case *ttlMinutes == 0:
		sandbox.ExpiresAt = nil
	default:
		if sandbox.ExpiresAt == nil {
			return sandbox, nil
		}
		newExpiry := sandbox.ExpiresAt.Add(time.Duration(*ttlMinutes) * time.Minute)
		if maxExpiry := time.Now().UTC().Add(s.cfg.MaxTTL); newExpiry.After(maxExpiry) {
			newExpiry = maxExpiry
		}
		sandbox.ExpiresAt = &newExpiry
	}

	if err := s.repo.Update(sandbox); err != nil {
		return nil, err
	}

	eventMeta := map[string]any{"ttlMinutes": ttlMinutes}
	if sandbox.ExpiresAt != nil {
		eventMeta["newExpiresAt"] = sandbox.ExpiresAt.Format(time.RFC3339)
	}
	_ = s.addEvent(sandbox.ID, "extended", eventMeta)
	resourceType := auditcontracts.ResourceTypeSandbox
	_ = s.audit.Log(AuditLogInput{
		Actor:        auditActor,
		Action:       auditcontracts.ActionSandboxTTLUpdated,
		ResourceType: &resourceType,
		ResourceID:   &sandbox.ID,
		Details: map[string]any{
			"ttlMinutes": ttlMinutes,
		},
	})

	return sandbox, nil
}

func (s *SandboxService) Delete(ctx context.Context, id uuid.UUID, auditActor AuditActor) error {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return ErrSandboxNotFound
	}

	resourceType := auditcontracts.ResourceTypeSandbox
	_ = s.audit.Log(AuditLogInput{
		Actor:        auditActor,
		Action:       auditcontracts.ActionSandboxDeleted,
		ResourceType: &resourceType,
		ResourceID:   &sandbox.ID,
		Details:      map[string]any{},
	})

	if sandbox.Status.IsActive() {
		_ = s.setStatus(sandbox, models.SandboxStatusStopping, strPtr("Container wird beendet"))
		go s.deleteContainerAsync(sandbox.ID, sandbox.ContainerID, sandbox.ImageID)
		return nil
	}

	return s.repo.DeleteByID(sandbox.ID)
}

func (s *SandboxService) deleteContainerAsync(sandboxID uuid.UUID, containerID string, imageID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	s.runPreStop(ctx, containerID, imageID)

	if err := s.docker.DeleteContainer(ctx, containerID); err != nil {
		_ = s.setStatusByID(sandboxID, models.SandboxStatusFailed, strPtr(fmt.Sprintf("Fehler beim Beenden: %v", err)))
		return
	}

	_ = s.setStatusByID(sandboxID, models.SandboxStatusDeleted, nil)
	_ = s.addEvent(sandboxID, "deleted", map[string]any{})
}

func (s *SandboxService) DeleteForGuest(ctx context.Context, id uuid.UUID, guestSessionID uuid.UUID, auditActor AuditActor) error {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return ErrSandboxNotFound
	}

	if sandbox.GuestSessionID == nil || *sandbox.GuestSessionID != guestSessionID {
		return ErrSandboxAccessDenied
	}

	return s.Delete(ctx, id, auditActor)
}

type CreateSnapshotInput struct {
	SandboxID   uuid.UUID
	Name        string
	Tag         string
	Title       *string
	Description *string
	IsPublic    bool
	ClientIP    string
	UserID      *uuid.UUID
	Metadata    json.RawMessage
	AuditActor  AuditActor
}

func (s *SandboxService) CreateSnapshot(ctx context.Context, input CreateSnapshotInput) (*models.Image, error) {
	targetImage := input.Name + ":" + input.Tag

	sandbox, err := s.repo.FindByID(input.SandboxID)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	var registryRef *string
	if sourceImage, srcErr := s.imageRepo.FindByID(sandbox.ImageID); srcErr == nil {
		ref := sourceImage.RegistryName()
		registryRef = &ref
	}

	image, err := s.images.CreateForCommit(
		input.UserID,
		input.Name,
		input.Tag,
		input.Title,
		input.Description,
		input.IsPublic,
		input.Metadata,
		registryRef,
	)
	if err != nil {
		return nil, err
	}

	_ = s.setStatus(sandbox, models.SandboxStatusPaused, strPtr("Snapshot wird erstellt"))

	_ = s.addEvent(sandbox.ID, "snapshotted", map[string]any{
		"targetImage": targetImage,
		"imageId":     image.ID.String(),
	})

	resourceType := auditcontracts.ResourceTypeImage
	_ = s.audit.Log(AuditLogInput{
		Actor:        input.AuditActor,
		Action:       auditcontracts.ActionImageSnapshotCreated,
		ResourceType: &resourceType,
		ResourceID:   &image.ID,
		Details: map[string]any{
			"sandboxId":   sandbox.ID.String(),
			"targetImage": targetImage,
		},
	})

	go s.commitSnapshot(sandbox.ID, sandbox.ContainerID, image.ID, targetImage)

	return image, nil
}

func (s *SandboxService) commitSnapshot(sandboxID uuid.UUID, containerID string, imageID uuid.UUID, targetImage string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	commitErr := s.docker.CommitContainer(ctx, containerID, targetImage)
	s.images.FinishCommit(imageID, commitErr)

	_ = s.setStatusByID(sandboxID, models.SandboxStatusRunning, nil)
}

func (s *SandboxService) ReconcileOnStartup(ctx context.Context) {
	stale, err := s.repo.ListByStatuses([]models.SandboxStatus{
		models.SandboxStatusStopping,
		models.SandboxStatusPaused,
	})
	if err != nil {
		slog.Error("reconcile: failed to list stale sandboxes", "component", "reconcile", "error", err.Error())
		return
	}
	for _, sb := range stale {
		if s.docker.ContainerExists(ctx, sb.ContainerID) {
			_ = s.setStatusByID(sb.ID, models.SandboxStatusRunning, nil)
			slog.Info("reconcile: restored sandbox to running", "component", "reconcile", "sandbox_id", sb.ID.String())
		} else {
			_ = s.setStatusByID(sb.ID, models.SandboxStatusFailed, strPtr("Vorgang durch API-Neustart unterbrochen"))
			slog.Info("reconcile: marked sandbox as failed", "component", "reconcile", "sandbox_id", sb.ID.String())
		}
	}
}

func (s *SandboxService) StartCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.CleanupInterval)
	slog.Info("sandbox cleanup loop started", "component", "cleanup", "interval", s.cfg.CleanupInterval.String())
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("sandbox cleanup loop stopped", "component", "cleanup")
				return
			case <-ticker.C:
				slog.Debug("running sandbox cleanup", "component", "cleanup")
				if err := s.CleanupExpired(ctx); err != nil {
					slog.Error("cleanup expired sandboxes failed", "component", "cleanup", "error", err.Error())
				}
			}
		}
	}()
}

func (s *SandboxService) StartDockerEventLoop(ctx context.Context) {
	events, errs := s.docker.SubscribeSandboxEvents(ctx)
	slog.Info("sandbox docker event loop started", "component", "sandbox_events")

	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("sandbox docker event loop stopped", "component", "sandbox_events")
				return
			case err, ok := <-errs:
				if !ok {
					return
				}
				if err != nil {
					slog.Error("sandbox docker event loop failed", "component", "sandbox_events", "error", err.Error())
				}
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				if err := s.handleDockerContainerEvent(event); err != nil {
					slog.Error("handle sandbox docker event failed",
						"component", "sandbox_events",
						"container_id", event.ContainerID,
						"action", event.Action,
						"error", err.Error(),
					)
				}
			}
		}
	}()
}

func (s *SandboxService) CleanupExpired(ctx context.Context) error {
	expired, err := s.repo.FindExpired(time.Now().UTC())
	if err != nil {
		return err
	}
	slog.Debug("expired sandboxes loaded", "component", "cleanup", "count", len(expired))

	// Expiration is database-driven so a process restart does not lose the
	// deletion schedule for previously created sandboxes.
	for _, sandbox := range expired {
		s.runPreStop(ctx, sandbox.ContainerID, sandbox.ImageID)

		if err := s.docker.DeleteContainer(ctx, sandbox.ContainerID); err != nil {
			slog.Error("delete expired container failed",
				"component", "cleanup",
				"sandbox_id", sandbox.ID.String(),
				"container_id", sandbox.ContainerID,
				"error", err.Error(),
			)
			continue
		}

		now := time.Now().UTC()
		expiredReason := "Laufzeit abgelaufen"
		sandbox.Status = models.SandboxStatusExpired
		sandbox.StateReason = &expiredReason
		sandbox.DeletedAt = &now
		if err := s.repo.Update(&sandbox); err != nil {
			slog.Error("update expired sandbox failed",
				"component", "cleanup",
				"sandbox_id", sandbox.ID.String(),
				"container_id", sandbox.ContainerID,
				"error", err.Error(),
			)
			continue
		}

		slog.Info("sandbox expired and cleaned up",
			"component", "cleanup",
			"sandbox_id", sandbox.ID.String(),
			"container_id", sandbox.ContainerID,
		)
		_ = s.addEvent(sandbox.ID, "expired", map[string]any{})
	}

	return nil
}

func (s *SandboxService) handleDockerContainerEvent(event docker.SandboxContainerEvent) error {
	sandbox, err := s.repo.FindByContainerID(event.ContainerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	switch event.Action {
	case "start":
		if sandbox.Status == models.SandboxStatusStopped {
			reason := "Container wird gestartet"
			sandbox.Status = models.SandboxStatusStarting
			sandbox.StateReason = &reason
			if err := s.repo.Update(sandbox); err != nil {
				return err
			}
			return s.addEvent(sandbox.ID, "started", map[string]any{
				"source": "docker_event",
				"action": event.Action,
			})
		}
		return nil

	case "pause":
		if sandbox.Status == models.SandboxStatusRunning {
			sandbox.Status = models.SandboxStatusPaused
			if sandbox.StateReason == nil {
				sandbox.StateReason = strPtr("Container pausiert")
			}
			if err := s.repo.Update(sandbox); err != nil {
				return err
			}
		}
		return nil

	case "unpause":
		if sandbox.Status == models.SandboxStatusPaused {
			sandbox.Status = models.SandboxStatusRunning
			sandbox.StateReason = nil
			if err := s.repo.Update(sandbox); err != nil {
				return err
			}
		}
		return nil

	case "stop", "die":
		if sandbox.Status == models.SandboxStatusDeleted || sandbox.Status == models.SandboxStatusExpired ||
			sandbox.Status == models.SandboxStatusStopping || sandbox.Status == models.SandboxStatusStarting {
			return nil
		}
		if sandbox.Status != models.SandboxStatusStopped {
			sandbox.Status = models.SandboxStatusStopped
			sandbox.StateReason = nil
			if err := s.repo.Update(sandbox); err != nil {
				return err
			}
			return s.addEvent(sandbox.ID, "stopped", map[string]any{
				"source": "docker_event",
				"action": event.Action,
			})
		}
		return nil

	case "destroy":
		if sandbox.Status == models.SandboxStatusDeleted || sandbox.Status == models.SandboxStatusExpired {
			return nil
		}
		sandbox.Status = models.SandboxStatusDeleted
		sandbox.StateReason = nil
		now := time.Now().UTC()
		sandbox.DeletedAt = &now
		if err := s.repo.Update(sandbox); err != nil {
			return err
		}
		return s.addEvent(sandbox.ID, "deleted", map[string]any{
			"source": "docker_event",
			"action": event.Action,
		})
	}

	return nil
}

func (s *SandboxService) enforceLimits(input CreateSandboxInput) error {
	total, err := s.repo.CountActiveTotal()
	if err != nil {
		return err
	}
	// The global guard protects the Docker host itself before user-specific
	// limits are evaluated.
	if total >= int64(s.guard.MaxActiveTotal) {
		return ErrSandboxLimitReached
	}

	if input.UserID == nil {
		// Guests are limited by IP because they do not have an employee account.
		ipCount, err := s.repo.CountActiveByIP(input.ClientIP)
		if err != nil {
			return err
		}
		if ipCount >= int64(s.guard.MaxPublicDemosPerIP) {
			return ErrSandboxLimitReached
		}
		return nil
	}

	userCount, err := s.repo.CountActiveByUser(*input.UserID)
	if err != nil {
		return err
	}
	// Employee limits intentionally bypass the guest IP limit because a team may
	// work behind the same office NAT.
	if userCount >= int64(s.guard.MaxActivePerUser) {
		return ErrSandboxLimitReached
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}

func (s *SandboxService) setStatus(sandbox *models.Sandbox, status models.SandboxStatus, reason *string) error {
	sandbox.Status = status
	sandbox.StateReason = reason
	return s.repo.Update(sandbox)
}

func (s *SandboxService) setStatusByID(id uuid.UUID, status models.SandboxStatus, reason *string) error {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	return s.setStatus(sandbox, status, reason)
}

func (s *SandboxService) addEvent(sandboxID uuid.UUID, eventType string, metadata map[string]any) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return s.eventRepo.Create(&models.SandboxEvent{
		ID:        uuid.New(),
		SandboxID: sandboxID,
		EventType: eventType,
		Metadata:  datatypes.JSON(payload),
		CreatedAt: time.Now().UTC(),
	})
}

func sandboxActorType(sandbox *models.Sandbox) string {
	if sandbox.OwnerID != nil {
		return "user"
	}
	return "guest"
}

// runPreStop resolves and executes pre stop lifecycle hooks for a container.
func (s *SandboxService) runPreStop(ctx context.Context, containerID string, imageID uuid.UUID) {
	registryName := s.imageRepo.ResolveRegistryName(imageID)
	if registryName == "" {
		return
	}

	resolved, err := s.resolver.Resolve(registryName, registry.TemplateContext{ImageName: registryName})
	if err != nil {
		slog.Warn("pre-stop resolve failed, skipping", "container_id", containerID, "image", registryName, "error", err)
		return
	}

	if len(resolved.PreStop) > 0 {
		preStopCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		s.executor.RunPreStop(preStopCtx, containerID, resolved.PreStop)
	}
}

// buildTemplateContext creates the registry template context for resolving environment variables, labels, and lifecycle commands
func (s *SandboxService) buildTemplateContext(
	imageName, containerName, hostname, sandboxID, ttl, expiresAt, clientIP, port string,
	metadata map[string]string,
) registry.TemplateContext {
	imageRepo, imageTag := splitImageRef(imageName)
	scheme := s.scheme()
	return registry.TemplateContext{
		Hostname:       hostname,
		URL:            scheme + "://" + hostname,
		Scheme:         scheme,
		Port:           port,
		ContainerName:  containerName,
		TrustedProxies: s.dockerCfg.TrustedProxies,
		DockerMode:     string(s.dockerCfg.Mode),
		Network:        s.dockerCfg.Network,
		InternalPort:   strconv.Itoa(s.cfg.InternalPort),
		ImageName:      imageName,
		ImageRepo:      imageRepo,
		ImageTag:       imageTag,
		SandboxID:      sandboxID,
		HostSuffix:     s.cfg.HostSuffix,
		TTL:            ttl,
		ExpiresAt:      expiresAt,
		ClientIP:       clientIP,
		Meta:           metadata,
	}
}

func (s *SandboxService) scheme() string {
	if s.dockerCfg.Mode == config.DockerModeTraefik && s.dockerCfg.TraefikCertResolver != "" {
		return "https"
	}
	return "http"
}

func splitImageRef(ref string) (string, string) {
	if strings.Contains(ref, "@") {
		return ref, ""
	}
	if i := strings.LastIndex(ref, ":"); i > strings.LastIndex(ref, "/") && i >= 0 {
		return ref[:i], ref[i+1:]
	}
	return ref, ""
}
