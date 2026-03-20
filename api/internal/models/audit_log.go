package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index" json:"userId,omitempty"`
	Action    string         `gorm:"size:128;not null;index" json:"action"`
	IPAddress string         `gorm:"size:128" json:"ipAddress"`
	Details   datatypes.JSON `gorm:"type:jsonb" json:"details" swaggertype:"object"`
	CreatedAt time.Time      `gorm:"not null;index" json:"createdAt"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
