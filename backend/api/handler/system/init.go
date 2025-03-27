package system

import "github.com/mr-pixel-kg/shopware-sandbox-plattform/services"

type SystemHandler struct {
	AuditLogService *services.AuditLogService
}

func NewSystemHandler(auditLogService *services.AuditLogService) *SystemHandler {
	return &SystemHandler{
		AuditLogService: auditLogService,
	}
}
