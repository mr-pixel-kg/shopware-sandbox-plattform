package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type SandboxEvent struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	SandboxID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"sandboxId"`
	EventType   string         `gorm:"size:64;not null" json:"eventType"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Metadata    datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt   time.Time      `gorm:"not null" json:"createdAt"`
}

func (SandboxEvent) TableName() string {
	return "sandbox_events"
}
