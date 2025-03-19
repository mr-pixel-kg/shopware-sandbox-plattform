package images

import (
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services/audit"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/services/images"
)

type ImageHandler struct {
	DockerService   *services.DockerService
	ImageService    *images.ImageService
	AuditLogService *audit.AuditLogService
}

func NewImageHandler(dockerService *services.DockerService, imageService *images.ImageService, auditLogService *audit.AuditLogService) *ImageHandler {
	return &ImageHandler{
		DockerService:   dockerService,
		ImageService:    imageService,
		AuditLogService: auditLogService,
	}
}
