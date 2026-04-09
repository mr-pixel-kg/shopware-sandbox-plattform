package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"gorm.io/datatypes"
)

type CreateSandboxRequest struct {
	ImageID     string            `json:"imageId" validate:"required,uuid" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	TTLMinutes  *int              `json:"ttlMinutes" validate:"omitempty,gte=0" example:"120"`
	DisplayName *string           `json:"displayName,omitempty" example:"My Test Shop"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type UpdateSandboxRequest struct {
	DisplayName *string `json:"displayName" example:"My Test Shop"`
	TTLMinutes  *int    `json:"ttlMinutes" validate:"omitempty,gte=0" example:"60"`
}

type CreateSnapshotRequest struct {
	ImagePayload
}

type SandboxListResponse struct {
	Data []SandboxResponse `json:"data"`
	Meta PaginatedMeta     `json:"meta"`
}

type SandboxResponse struct {
	ID            uuid.UUID            `json:"id" format:"uuid" example:"0b443c82-d8a3-49a7-b59a-26ce327c7341"`
	ImageID       uuid.UUID            `json:"imageId" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Owner         *UserSummary         `json:"owner,omitempty"`
	ClientID      *uuid.UUID           `json:"clientId,omitempty" format:"uuid" example:"db7fcb92-c2ff-4c20-9ac2-5a2504ab6326"`
	DisplayName   string               `json:"displayName" example:"My Test Shop"`
	Status        models.SandboxStatus `json:"status" enums:"starting,running,paused,stopping,stopped,expired,deleted,failed" example:"running"`
	StateReason   *string              `json:"stateReason,omitempty" example:"Snapshot wird erstellt"`
	ContainerID   string               `json:"containerId" example:"1a2b3c4d5e6f7g8h9i0j"`
	ContainerName string               `json:"containerName" example:"sandbox-0b443c82"`
	URL           string               `json:"url" example:"https://sandbox-0b443c82.demo.shopshredder.de"`
	Port          *int                 `json:"port,omitempty" example:"8080"`
	SSH           *SSHConnectionInfo   `json:"ssh,omitempty"`
	ClientIP      string               `json:"clientIp" example:"203.0.113.25"`
	Metadata      datatypes.JSON       `json:"metadata,omitempty" swaggertype:"string"`
	ExpiresAt     *time.Time           `json:"expiresAt,omitempty" example:"2026-03-20T12:00:00Z"`
	LastSeenAt    *time.Time           `json:"lastSeenAt,omitempty" example:"2026-03-20T10:45:00Z"`
	CreatedAt     time.Time            `json:"createdAt" example:"2026-03-20T10:15:00Z"`
	UpdatedAt     time.Time            `json:"updatedAt" example:"2026-03-20T10:20:00Z"`
}

type SSHConnectionInfo struct {
	Host     string `json:"host" example:"sandbox-abc.zion.mr-pixel.de"`
	Port     int    `json:"port" example:"2222"`
	Username string `json:"username" example:"aiomayo+0b443c82-d8a3-49a7-b59a-26ce327c7341"`
	Password string `json:"password" example:"test123123"`
	Command  string `json:"command" example:"ssh aiomayo+0b443c82@sandbox-abc.zion.mr-pixel.de -p 2222"`
}

type SandboxStreamEvent struct {
	ID          uuid.UUID `json:"id" format:"uuid" example:"0b443c82-d8a3-49a7-b59a-26ce327c7341"`
	Status      string    `json:"status" example:"starting"`
	StateReason string    `json:"stateReason,omitempty" example:"Container wird gestartet"`
}

type SandboxHealthEvent struct {
	SandboxID     uuid.UUID `json:"sandboxId" format:"uuid" example:"0b443c82-d8a3-49a7-b59a-26ce327c7341"`
	Status        string    `json:"status" example:"probing"`
	Ready         bool      `json:"ready" example:"false"`
	URL           string    `json:"url" example:"https://sandbox-0b443c82.demo.shopshredder.de"`
	HTTPStatus    int       `json:"httpStatus,omitempty" example:"200"`
	LatencyMs     int64     `json:"latencyMs,omitempty" example:"412"`
	FailureReason string    `json:"failureReason,omitempty" example:"tls_handshake_failed"`
	Message       string    `json:"message,omitempty" example:"Sandbox URL is reachable"`
	CheckedAt     time.Time `json:"checkedAt" example:"2026-03-23T10:15:07Z"`
}
