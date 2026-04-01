package services

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"gorm.io/datatypes"
)

type AuditService struct {
	repo *repositories.AuditLogRepository
}

type AuditActor struct {
	UserID      *uuid.UUID
	IPAddress   *string
	UserAgent   *string
	ClientToken *uuid.UUID
}

type AuditLogInput struct {
	Actor        AuditActor
	Action       auditcontracts.Action
	ResourceType *auditcontracts.ResourceType
	ResourceID   *uuid.UUID
	Details      map[string]any
}

type AuditLogListInput struct {
	Limit        int
	Offset       int
	UserID       *uuid.UUID
	Action       *string
	ResourceType *string
	ResourceID   *uuid.UUID
	ClientToken  *uuid.UUID
	From         *time.Time
	To           *time.Time
}

type AuditLogListResult struct {
	Logs   []models.AuditLog
	Total  int64
	Limit  int
	Offset int
}

func NewAuditService(repo *repositories.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(input AuditLogInput) error {
	details := input.Details
	if details == nil {
		details = map[string]any{}
	}

	payload, err := json.Marshal(details)
	if err != nil {
		return err
	}

	var resourceType *string
	if input.ResourceType != nil {
		value := string(*input.ResourceType)
		resourceType = &value
	}

	return s.repo.Create(&models.AuditLog{
		ID:           uuid.New(),
		UserID:       input.Actor.UserID,
		Action:       string(input.Action),
		IPAddress:    input.Actor.IPAddress,
		UserAgent:    input.Actor.UserAgent,
		ClientToken:  input.Actor.ClientToken,
		ResourceType: resourceType,
		ResourceID:   input.ResourceID,
		Details:      datatypes.JSON(payload),
		Timestamp:    time.Now().UTC(),
	})
}

func (s *AuditService) List(input AuditLogListInput) (*AuditLogListResult, error) {
	if input.Limit <= 0 {
		input.Limit = 50
	}
	if input.Limit > 500 {
		input.Limit = 500
	}
	if input.Offset < 0 {
		input.Offset = 0
	}
	if input.Action != nil {
		value := strings.TrimSpace(*input.Action)
		if value == "" {
			input.Action = nil
		} else {
			input.Action = &value
		}
	}
	if input.ResourceType != nil {
		value := strings.TrimSpace(*input.ResourceType)
		if value == "" {
			input.ResourceType = nil
		} else {
			input.ResourceType = &value
		}
	}

	logs, total, err := s.repo.List(repositories.AuditLogListOptions{
		Limit:        input.Limit,
		Offset:       input.Offset,
		UserID:       input.UserID,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		ClientToken:  input.ClientToken,
		From:         input.From,
		To:           input.To,
	})
	if err != nil {
		return nil, err
	}

	return &AuditLogListResult{
		Logs:   logs,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}, nil
}
