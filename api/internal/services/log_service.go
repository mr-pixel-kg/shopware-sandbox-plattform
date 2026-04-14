package services

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/docker"
	"github.com/mr-pixel-kg/shopshredder/api/internal/lifecycle"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"github.com/mr-pixel-kg/shopshredder/api/internal/registry"
	"github.com/mr-pixel-kg/shopshredder/api/internal/repositories"
)

var (
	ErrLogAccessDenied   = errors.New("log access denied")
	ErrLogNotActive      = errors.New("sandbox is not active")
	ErrLogSourceNotFound = errors.New("log source not found")
)

type LogService struct {
	docker    docker.Client
	repo      *repositories.SandboxRepository
	imgRepo   *repositories.ImageRepository
	resolver  *registry.Resolver
	lifecycle *lifecycle.Store
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
	lifecycleStore *lifecycle.Store,
) *LogService {
	return &LogService{
		docker:    dockerClient,
		repo:      repo,
		imgRepo:   imgRepo,
		resolver:  resolver,
		lifecycle: lifecycleStore,
	}
}

func (s *LogService) ValidateAccess(input ValidateLogAccessInput) (*models.Sandbox, error) {
	sandbox, err := s.repo.FindByID(input.SandboxID)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	if !sandbox.Status.IsActive() {
		return nil, ErrLogNotActive
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

	sources := make([]registry.LogSource, 0, len(entry.Logs)+1)

	if len(entry.PostStart) > 0 || len(entry.PreStop) > 0 {
		sources = append(sources, registry.LogSource{
			Key:   "lifecycle",
			Label: "Lifecycle",
			Type:  registry.LogSourceTypeLifecycle,
		})
	}

	sources = append(sources, entry.Logs...)
	return sources
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

type StreamLogOptions struct {
	Verbose bool
}

func (s *LogService) StreamLog(ctx context.Context, containerID string, source registry.LogSource, opts StreamLogOptions) (*docker.LogStream, error) {
	switch source.Type {
	case registry.LogSourceTypeDocker:
		return s.docker.ContainerLogs(ctx, containerID)
	case registry.LogSourceTypeFile:
		reader, err := s.docker.ExecFollow(ctx, containerID, []string{"tail", "-f", source.Path})
		if err != nil {
			return nil, err
		}
		return &docker.LogStream{Reader: reader, TTY: false}, nil
	case registry.LogSourceTypeLifecycle:
		reader := s.streamLifecycle(ctx, containerID, opts.Verbose)
		return &docker.LogStream{Reader: reader, TTY: true}, nil
	default:
		return nil, fmt.Errorf("unsupported log source type: %s", source.Type)
	}
}

func (s *LogService) streamLifecycle(ctx context.Context, containerID string, verbose bool) io.ReadCloser {
	buf := s.lifecycle.GetOrCreate(containerID)
	snapshot, ch, cancel := buf.SnapshotAndSubscribe()

	pr, pw := io.Pipe()

	go func() {
		defer func() { _ = pw.Close() }()
		defer cancel()

		for _, e := range snapshot {
			if !verbose && lifecycle.IsVerbose(e.Level) {
				continue
			}
			if _, err := fmt.Fprintln(pw, lifecycle.FormatEntry(e)); err != nil {
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case entry, ok := <-ch:
				if !ok {
					return
				}
				if !verbose && lifecycle.IsVerbose(entry.Level) {
					continue
				}
				if _, err := fmt.Fprintln(pw, lifecycle.FormatEntry(entry)); err != nil {
					return
				}
			}
		}
	}()

	return pr
}
