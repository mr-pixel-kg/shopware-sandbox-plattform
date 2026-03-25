package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SandboxStatus string

const (
	SandboxStatusStarting SandboxStatus = "starting"
	SandboxStatusRunning  SandboxStatus = "running"
	SandboxStatusStopped  SandboxStatus = "stopped"
	SandboxStatusExpired  SandboxStatus = "expired"
	SandboxStatusDeleted  SandboxStatus = "deleted"
	SandboxStatusFailed   SandboxStatus = "failed"
)

type Sandbox struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"0b443c82-d8a3-49a7-b59a-26ce327c7341"`
	ImageID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"imageId" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	CreatedByUserID *uuid.UUID     `gorm:"type:uuid;index" json:"createdByUserId,omitempty" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	GuestSessionID  *uuid.UUID     `gorm:"type:uuid;index" json:"guestSessionId,omitempty" format:"uuid" example:"db7fcb92-c2ff-4c20-9ac2-5a2504ab6326"`
	Status          SandboxStatus  `gorm:"size:32;not null;index" json:"status" enums:"starting,running,stopped,expired,deleted,failed" example:"running"`
	ContainerID     string         `gorm:"size:255;not null;uniqueIndex" json:"containerId" example:"1a2b3c4d5e6f7g8h9i0j"`
	ContainerName   string         `gorm:"size:255;not null;uniqueIndex" json:"containerName" example:"sandbox-0b443c82"`
	URL             string         `gorm:"size:1024;not null;uniqueIndex" json:"url" example:"https://sandbox-0b443c82.demo.shopshredder.de"`
	Port            *int           `gorm:"default:null" json:"port,omitempty" example:"8080"`
	ClientIP        string         `gorm:"size:128;not null;index" json:"clientIp" example:"203.0.113.25"`
	Metadata        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty" swaggertype:"string"`
	ExpiresAt       *time.Time     `gorm:"index" json:"expiresAt,omitempty" example:"2026-03-20T12:00:00Z"`
	LastSeenAt      *time.Time     `json:"lastSeenAt,omitempty" example:"2026-03-20T10:45:00Z"`
	BaseModel
}

func (Sandbox) TableName() string {
	return "sandboxes"
}
