package services

import (
	"encoding/json"
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
		CreatedAt:    time.Now().UTC(),
	})
}

func (s *AuditService) List(limit int) ([]models.AuditLog, error) {
	return s.repo.List(limit)
}
