package services

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/models"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/repository"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/middleware"
	"log"
)

type AuditLogService struct {
	AuditLogRepository *repository.AuditLogRepository
}

func NewAuditLogService(auditLogRepository *repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{AuditLogRepository: auditLogRepository}
}

func (s *AuditLogService) Log(ip, userAgent string, username *string, action models.AuditAction, details map[string]interface{}) error {

	// Check if details are nil and use empty map if necessary
	if details == nil {
		details = make(map[string]interface{})
	}

	detailsJson, err := json.Marshal(details)
	if err != nil {
		log.Printf("Failed to serialize details: %v", err)
		return err
	}

	err = s.AuditLogRepository.Create(&models.AuditLogEntry{
		IpAddress: ip,
		UserAgent: userAgent,
		Username:  username,
		Action:    action,
		Details:   detailsJson,
	})
	if err != nil {
		log.Printf("Failed to store audit log entry: %v", err)
		return err
	}

	return err
}

func (s *AuditLogService) LogRequest(ctx echo.Context, action models.AuditAction, details map[string]interface{}) error {
	ip := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	if !middleware.IsUserLoggedIn(ctx) {
		// If user is not logged in
		return s.Log(ip, userAgent, nil, action, details)
	}

	username := middleware.GetCurrentUserName(ctx)
	return s.Log(ip, userAgent, username, action, details)
}
