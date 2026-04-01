package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
)

var (
	ErrTerminalSessionLimit = errors.New("terminal session limit reached")
	ErrTerminalNotRunning   = errors.New("sandbox is not running")
	ErrTerminalAccessDenied = errors.New("terminal access denied")
)

type TerminalService struct {
	cfg      config.TerminalConfig
	docker   docker.Client
	repo     *repositories.SandboxRepository
	sessions sync.Map
}

type ValidateTerminalAccessInput struct {
	SandboxID uuid.UUID
	UserID    uuid.UUID
	IsAdmin   bool
}

func NewTerminalService(
	cfg config.TerminalConfig,
	dockerClient docker.Client,
	repo *repositories.SandboxRepository,
) *TerminalService {
	return &TerminalService{
		cfg:    cfg,
		docker: dockerClient,
		repo:   repo,
	}
}

func (s *TerminalService) Config() config.TerminalConfig {
	return s.cfg
}

func (s *TerminalService) ValidateAccess(input ValidateTerminalAccessInput) (*models.Sandbox, error) {
	sandbox, err := s.repo.FindByID(input.SandboxID)
	if err != nil {
		return nil, ErrSandboxNotFound
	}

	if sandbox.Status != models.SandboxStatusRunning {
		return nil, ErrTerminalNotRunning
	}

	if input.IsAdmin {
		return sandbox, nil
	}

	if sandbox.OwnerID == nil || *sandbox.OwnerID != input.UserID {
		return nil, ErrTerminalAccessDenied
	}

	return sandbox, nil
}

func (s *TerminalService) OpenSession(ctx context.Context, sandbox *models.Sandbox, cols, rows uint) (*docker.ExecSession, error) {
	counter := s.getOrCreateCounter(sandbox.ID)
	current := atomic.AddInt32(counter, 1)
	if int(current) > s.cfg.MaxSessionsPerSandbox {
		atomic.AddInt32(counter, -1)
		return nil, ErrTerminalSessionLimit
	}

	session, err := s.docker.CreateExecSession(ctx, sandbox.ContainerID, docker.ExecAttachOptions{
		Cols: cols,
		Rows: rows,
	})
	if err != nil {
		atomic.AddInt32(counter, -1)
		return nil, fmt.Errorf("create exec session: %w", err)
	}

	return session, nil
}

func (s *TerminalService) CloseSession(sandboxID uuid.UUID) {
	if counter := s.getCounter(sandboxID); counter != nil {
		if atomic.AddInt32(counter, -1) <= 0 {
			s.sessions.Delete(sandboxID)
		}
	}
}

func (s *TerminalService) ActiveSessions(sandboxID uuid.UUID) int {
	if counter := s.getCounter(sandboxID); counter != nil {
		return int(atomic.LoadInt32(counter))
	}
	return 0
}

func (s *TerminalService) getOrCreateCounter(sandboxID uuid.UUID) *int32 {
	val, _ := s.sessions.LoadOrStore(sandboxID, new(int32))
	return val.(*int32)
}

func (s *TerminalService) getCounter(sandboxID uuid.UUID) *int32 {
	val, ok := s.sessions.Load(sandboxID)
	if !ok {
		return nil
	}
	return val.(*int32)
}
