package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"gorm.io/datatypes"
)

var (
	ErrSandboxLimitReached = errors.New("sandbox limit reached")
	ErrSandboxNotFound     = errors.New("sandbox not found")
)

type SandboxService struct {
	cfg       config.SandboxConfig
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
}

func (s *SandboxService) ListActive() ([]models.Sandbox, error) {
	return s.repo.ListAllActive()
}

func (s *SandboxService) ListByUser(userID uuid.UUID) ([]models.Sandbox, error) {
	return s.repo.ListActiveByUser(userID)
}

func (s *SandboxService) ListByGuestSession(sessionID uuid.UUID) ([]models.Sandbox, error) {
	return s.repo.ListActiveByGuestSession(sessionID)
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

	image, err := s.imageRepo.FindByID(input.ImageID)
	if err != nil {
		return nil, err
	}
	if err := s.docker.EnsureImage(ctx, image.FullName()); err != nil {
		return nil, err
	}

	sandboxID := uuid.New()
	containerName := fmt.Sprintf("%s%s", s.cfg.URLPrefix, sandboxID.String())
	hostname := fmt.Sprintf("%s%s", containerName, s.cfg.HostSuffix)
	ttl := s.cfg.DefaultTTL
	if input.TTL != nil {
		ttl = *input.TTL
	}
	if ttl > s.cfg.MaxTTL {
		ttl = s.cfg.MaxTTL
	}

	container, err := s.docker.CreateContainer(ctx, docker.SandboxCreateRequest{
		ImageName:     image.FullName(),
		ContainerName: containerName,
		Hostname:      hostname,
	})
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(ttl)
	sandbox := &models.Sandbox{
		ID:              sandboxID,
		ImageID:         image.ID,
		CreatedByUserID: input.UserID,
		GuestSessionID:  input.GuestSessionID,
		Status:          models.SandboxStatusRunning,
		ContainerID:     container.ID,
		ContainerName:   container.Name,
		URL:             "https://" + hostname,
		ClientIP:        input.ClientIP,
		ExpiresAt:       &expiresAt,
	}

	if err := s.repo.Create(sandbox); err != nil {
		return nil, err
	}

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

func (s *SandboxService) Delete(ctx context.Context, id uuid.UUID, clientIP string, userID *uuid.UUID) error {
	sandbox, err := s.repo.FindByID(id)
	if err != nil {
		return ErrSandboxNotFound
	}

	if err := s.docker.DeleteContainer(ctx, sandbox.ContainerID); err != nil {
		return err
	}

	now := time.Now().UTC()
	sandbox.Status = models.SandboxStatusDeleted
	sandbox.DeletedAt = &now
	if err := s.repo.Update(sandbox); err != nil {
		return err
	}

	if err := s.addEvent(sandbox.ID, "deleted", map[string]any{}); err != nil {
		return err
	}

	_ = s.audit.Log(userID, "sandbox.deleted", clientIP, map[string]any{
		"sandboxId": sandbox.ID.String(),
	})

	return nil
}

type CreateSnapshotInput struct {
	SandboxID    uuid.UUID
	Name         string
	Tag          string
	Title        *string
	Description  *string
	ThumbnailURL *string
	IsPublic     bool
	ClientIP     string
	UserID       *uuid.UUID
}

func (s *SandboxService) CreateSnapshot(ctx context.Context, input CreateSnapshotInput) (*models.Image, error) {
	targetImage := input.Name + ":" + input.Tag

	sandbox, err := s.repo.FindByID(input.SandboxID)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	if err := s.docker.CommitContainer(ctx, sandbox.ContainerID, targetImage); err != nil {
		return nil, err
	}

	image, err := s.images.CreateForUser(
		ctx,
		input.UserID,
		input.Name,
		input.Tag,
		input.Title,
		input.Description,
		input.ThumbnailURL,
		input.IsPublic,
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
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.CleanupExpired(ctx); err != nil {
					log.Printf("cleanup expired sandboxes: %v", err)
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

	for _, sandbox := range expired {
		if err := s.docker.DeleteContainer(ctx, sandbox.ContainerID); err != nil {
			log.Printf("delete expired container %s: %v", sandbox.ContainerID, err)
			continue
		}

		now := time.Now().UTC()
		sandbox.Status = models.SandboxStatusExpired
		sandbox.DeletedAt = &now
		if err := s.repo.Update(&sandbox); err != nil {
			log.Printf("update expired sandbox %s: %v", sandbox.ID, err)
			continue
		}

		_ = s.addEvent(sandbox.ID, "expired", map[string]any{})
	}

	return nil
}

func (s *SandboxService) enforceLimits(input CreateSandboxInput) error {
	total, err := s.repo.CountActiveTotal()
	if err != nil {
		return err
	}
	if total >= int64(s.guard.MaxActiveTotal) {
		return ErrSandboxLimitReached
	}

	if input.UserID == nil {
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
	if userCount >= int64(s.guard.MaxActivePerUser) {
		return ErrSandboxLimitReached
	}

	return nil
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
	if sandbox.CreatedByUserID != nil {
		return "user"
	}
	return "guest"
}
