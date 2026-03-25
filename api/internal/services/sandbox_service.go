package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
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
	}
}

type CreateSandboxInput struct {
	ImageID        uuid.UUID
	UserID         *uuid.UUID
	GuestSessionID *uuid.UUID
	ClientIP       string
	TTL            *time.Duration
	Metadata       map[string]string
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
		return nil, fmt.Errorf("image pull failed: %s", ptrStr(image.Error))
	}
	if err := s.docker.EnsureImage(ctx, image.FullName()); err != nil {
		return nil, err
	}

	sandboxID := uuid.New()
	containerName := fmt.Sprintf("%s%s", s.cfg.URLPrefix, sandboxID.String())
	ttl := s.cfg.DefaultTTL
	if input.TTL != nil {
		ttl = *input.TTL
	}
	// Clamp requested lifetimes to the configured maximum to keep cleanup
	// predictable even for authenticated users.
	if ttl > s.cfg.MaxTTL {
		ttl = s.cfg.MaxTTL
	}

	var hostname string
	if s.dockerCfg.Mode == config.DockerModeTraefik {
		hostname = fmt.Sprintf("%s%s", containerName, s.cfg.HostSuffix)
	}

	expiresAt := time.Now().UTC().Add(ttl)
	container, err := s.docker.CreateContainer(ctx, docker.SandboxCreateRequest{
		ImageName:     image.FullName(),
		RegistryRef:   image.RegistryName(),
		ContainerName: containerName,
		Hostname:      hostname,
		SandboxID:     sandboxID.String(),
		TTL:           ttl.String(),
		ExpiresAt:     expiresAt.Format(time.RFC3339),
		ClientIP:      input.ClientIP,
		Metadata:      input.Metadata,
	})
	if err != nil {
		return nil, err
	}

	fieldsJSON, _ := json.Marshal(input.Metadata)
	sandbox := &models.Sandbox{
		ID:              sandboxID,
		ImageID:         image.ID,
		CreatedByUserID: input.UserID,
		GuestSessionID:  input.GuestSessionID,
		Status:          models.SandboxStatusStarting,
		ContainerID:     container.ID,
		ContainerName:   container.Name,
		URL:             container.URL,
		Port:            container.Port,
		ClientIP:        input.ClientIP,
		Metadata:        datatypes.JSON(fieldsJSON),
		ExpiresAt:       &expiresAt,
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

	_ = s.audit.Log(input.UserID, "sandbox.created", input.ClientIP, map[string]any{
		"sandboxId": sandbox.ID.String(),
		"imageId":   image.ID.String(),
		"actor":     sandboxActorType(sandbox),
	})

	return sandbox, nil
}

func (s *SandboxService) ExtendTTL(id uuid.UUID, additionalMinutes int, clientIP string, userID *uuid.UUID) (*models.Sandbox, error) {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	if sandbox.Status != models.SandboxStatusRunning && sandbox.Status != models.SandboxStatusStarting {
		return nil, ErrSandboxNotFound
	}

	if sandbox.ExpiresAt == nil {
		return nil, ErrSandboxNotFound
	}

	additional := time.Duration(additionalMinutes) * time.Minute
	newExpiry := sandbox.ExpiresAt.Add(additional)

	maxExpiry := time.Now().UTC().Add(s.cfg.MaxTTL)
	if newExpiry.After(maxExpiry) {
		newExpiry = maxExpiry
	}

	sandbox.ExpiresAt = &newExpiry
	if err := s.repo.Update(sandbox); err != nil {
		return nil, err
	}

	_ = s.addEvent(sandbox.ID, "extended", map[string]any{
		"additionalMinutes": additionalMinutes,
		"newExpiresAt":      newExpiry.Format(time.RFC3339),
	})

	_ = s.audit.Log(userID, "sandbox.extended", clientIP, map[string]any{
		"sandboxId":         sandbox.ID.String(),
		"additionalMinutes": additionalMinutes,
	})

	return sandbox, nil
}

func (s *SandboxService) Delete(ctx context.Context, id uuid.UUID, clientIP string, userID *uuid.UUID) error {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return ErrSandboxNotFound
	}

	isActive := sandbox.Status == models.SandboxStatusRunning || sandbox.Status == models.SandboxStatusStarting

	if isActive {
		// Soft delete: stop container, mark as deleted, keep in history.
		if err := s.docker.DeleteContainer(ctx, sandbox.ContainerID, s.resolveImageName(sandbox.ImageID)); err != nil {
			return err
		}

		sandbox.Status = models.SandboxStatusDeleted
		if err := s.repo.Update(sandbox); err != nil {
			return err
		}

		if err := s.addEvent(sandbox.ID, "deleted", map[string]any{}); err != nil {
			return err
		}
	} else {
		// Hard delete: permanently remove from history.
		if err := s.repo.DeleteByID(sandbox.ID); err != nil {
			return err
		}
	}

	_ = s.audit.Log(userID, "sandbox.deleted", clientIP, map[string]any{
		"sandboxId": sandbox.ID.String(),
	})

	return nil
}

func (s *SandboxService) DeleteForGuest(ctx context.Context, id uuid.UUID, guestSessionID uuid.UUID, clientIP string) error {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return ErrSandboxNotFound
	}

	if sandbox.GuestSessionID == nil || *sandbox.GuestSessionID != guestSessionID {
		return ErrSandboxAccessDenied
	}

	return s.Delete(ctx, id, clientIP, nil)
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
}

func (s *SandboxService) CreateSnapshot(ctx context.Context, input CreateSnapshotInput) (*models.Image, error) {
	targetImage := input.Name + ":" + input.Tag

	sandbox, err := s.repo.FindByID(input.SandboxID)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	// use a separate context so that when docker clinet disconnects because of a timeout it doesnt cancels the potentially long image commit
	commitCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := s.docker.CommitContainer(commitCtx, sandbox.ContainerID, targetImage); err != nil {
		return nil, err
	}

	var registryRef *string
	if sourceImage, srcErr := s.imageRepo.FindByID(sandbox.ImageID); srcErr == nil {
		ref := sourceImage.RegistryName()
		registryRef = &ref
	}

	image, err := s.images.CreateForUser(
		ctx,
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

	if err := s.addEvent(sandbox.ID, "snapshotted", map[string]any{
		"targetImage": targetImage,
		"imageId":     image.ID.String(),
	}); err != nil {
		return nil, err
	}

	_ = s.audit.Log(input.UserID, "sandbox.snapshotted", input.ClientIP, map[string]any{
		"sandboxId":   sandbox.ID.String(),
		"targetImage": targetImage,
		"imageId":     image.ID.String(),
	})

	return image, nil
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
		if err := s.docker.DeleteContainer(ctx, sandbox.ContainerID, s.resolveImageName(sandbox.ImageID)); err != nil {
			slog.Error("delete expired container failed",
				"component", "cleanup",
				"sandbox_id", sandbox.ID.String(),
				"container_id", sandbox.ContainerID,
				"error", err.Error(),
			)
			continue
		}

		now := time.Now().UTC()
		sandbox.Status = models.SandboxStatusExpired
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
			sandbox.Status = models.SandboxStatusStarting
			if err := s.repo.Update(sandbox); err != nil {
				return err
			}
			return s.addEvent(sandbox.ID, "started", map[string]any{
				"source": "docker_event",
				"action": event.Action,
			})
		}
		return nil

	case "stop", "die":
		if sandbox.Status == models.SandboxStatusDeleted || sandbox.Status == models.SandboxStatusExpired {
			return nil
		}
		if sandbox.Status != models.SandboxStatusStopped {
			sandbox.Status = models.SandboxStatusStopped
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

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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

func (s *SandboxService) resolveImageName(imageID uuid.UUID) string {
	img, err := s.imageRepo.FindByID(imageID)
	if err != nil {
		return ""
	}
	return img.FullName()
}

func sandboxActorType(sandbox *models.Sandbox) string {
	if sandbox.CreatedByUserID != nil {
		return "user"
	}
	return "guest"
}
