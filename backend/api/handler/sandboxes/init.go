package sandboxes

import (
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services"
)

type SandboxHandler struct {
	SandboxService  *services.SandboxService
	AuditLogService *services.AuditLogService
	GuardService    *services.GuardService
}

func NewSandboxHandler(sandboxService *services.SandboxService, auditLogService *services.AuditLogService, guardService *services.GuardService) *SandboxHandler {
	return &SandboxHandler{
		SandboxService:  sandboxService,
		AuditLogService: auditLogService,
		GuardService:    guardService,
	}
}
