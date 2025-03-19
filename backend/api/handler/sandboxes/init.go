package sandboxes

import (
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services/audit"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services/sandbox"
)

type SandboxHandler struct {
	SandboxService  *sandbox.SandboxService
	AuditLogService *audit.AuditLogService
}

func NewSandboxHandler(sandboxService *sandbox.SandboxService, auditLogService *audit.AuditLogService) *SandboxHandler {
	return &SandboxHandler{
		SandboxService:  sandboxService,
		AuditLogService: auditLogService,
	}
}
