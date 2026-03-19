package models

import (
	"time"

	"github.com/google/uuid"
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
	ID              uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	ImageID         uuid.UUID     `gorm:"type:uuid;not null;index" json:"imageId"`
	CreatedByUserID *uuid.UUID    `gorm:"type:uuid;index" json:"createdByUserId,omitempty"`
	GuestSessionID  *uuid.UUID    `gorm:"type:uuid;index" json:"guestSessionId,omitempty"`
	Status          SandboxStatus `gorm:"size:32;not null;index" json:"status"`
	ContainerID     string        `gorm:"size:255;not null;uniqueIndex" json:"containerId"`
	ContainerName   string        `gorm:"size:255;not null;uniqueIndex" json:"containerName"`
	URL             string        `gorm:"size:1024;not null;uniqueIndex" json:"url"`
	Port            int           `gorm:"default:0" json:"port,omitempty"`
	ClientIP        string        `gorm:"size:128;not null;index" json:"clientIp"`
	ExpiresAt       *time.Time    `gorm:"index" json:"expiresAt,omitempty"`
	LastSeenAt      *time.Time    `json:"lastSeenAt,omitempty"`
	BaseModel
}

func (Sandbox) TableName() string {
	return "sandboxes"
}
