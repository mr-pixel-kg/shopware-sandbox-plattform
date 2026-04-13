package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/docker"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"github.com/mr-pixel-kg/shopshredder/api/internal/registry"
	"github.com/mr-pixel-kg/shopshredder/api/internal/repositories"
)

var (
	ErrLogAccessDenied   = errors.New("log access denied")
	ErrLogNotRunning     = errors.New("sandbox is not running")
	ErrLogSourceNotFound = errors.New("log source not found")
)

type LogService struct {
	docker   docker.Client
	repo     *repositories.SandboxRepository
	imgRepo  *repositories.ImageRepository
	resolver *registry.Resolver
}

type ValidateLogAccessInput struct {
	SandboxID uuid.UUID
	UserID    uuid.UUID
	IsAdmin   bool
}

func NewLogService(
	dockerClient docker.Client,
	repo *repositories.SandboxRepository,
	imgRepo *repositories.ImageRepository,
	resolver *registry.Resolver,
) *LogService {
	return &LogService{
		docker:   dockerClient,
		repo:     repo,
		imgRepo:  imgRepo,
		resolver: resolver,
	}
}

func (s *LogService) ValidateAccess(input ValidateLogAccessInput) (*models.Sandbox, error) {
	sandbox, err := s.repo.FindByID(input.SandboxID)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	if sandbox.Status != models.SandboxStatusRunning {
		return nil, ErrLogNotRunning
	}

	if input.IsAdmin {
		return sandbox, nil
	}

	if sandbox.OwnerID == nil || *sandbox.OwnerID != input.UserID {
		return nil, ErrLogAccessDenied
	}

	return sandbox, nil
}

func (s *LogService) GetLogSources(sandbox *models.Sandbox) []registry.LogSource {
	img, err := s.imgRepo.FindByID(sandbox.ImageID)
	if err != nil {
		return nil
	}
	entry := s.resolver.ResolveEntry(img.RegistryName())
	if entry == nil {
		return nil
	}
	return entry.Logs
}

func (s *LogService) FindLogSource(sandbox *models.Sandbox, key string) (*registry.LogSource, error) {
	sources := s.GetLogSources(sandbox)
	for _, src := range sources {
		if src.Key == key {
			return &src, nil
		}
	}
	return nil, ErrLogSourceNotFound
}

func (s *LogService) StreamLog(ctx context.Context, containerID string, source registry.LogSource) (*docker.LogStream, error) {
	switch source.Type {
	case registry.LogSourceTypeDocker:
		return s.docker.ContainerLogs(ctx, containerID)
	case registry.LogSourceTypeFile:
		reader, err := s.docker.ExecFollow(ctx, containerID, []string{"tail", "-f", source.Path})
		if err != nil {
			return nil, err
		}
		return &docker.LogStream{Reader: reader, TTY: false}, nil
	default:
		return nil, fmt.Errorf("unsupported log source type: %s", source.Type)
	}
}
