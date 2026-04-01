package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"4d0dbf0d-1034-42ef-8b6d-7eb3ceef99cf"`
	UserID       *uuid.UUID     `gorm:"type:uuid;index" json:"userId" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	User         *User          `gorm:"foreignKey:UserID;references:ID" json:"-" swaggerignore:"true"`
	Action       string         `gorm:"size:128;not null;index" json:"action" example:"sandbox.created"`
	IPAddress    *string        `gorm:"size:128" json:"ipAddress" example:"203.0.113.25"`
	UserAgent    *string        `gorm:"size:512" json:"userAgent" example:"Mozilla/5.0"`
	ClientToken  *uuid.UUID     `gorm:"type:uuid;index" json:"clientToken" format:"uuid" example:"4d0dbf0d-1034-42ef-8b6d-7eb3ceef99cf"`
	ResourceType *string        `gorm:"size:64;index" json:"resourceType" example:"sandbox"`
	ResourceID   *uuid.UUID     `gorm:"type:uuid;index" json:"resourceId" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	Details      datatypes.JSON `gorm:"type:jsonb" json:"details" swaggertype:"object"`
	Timestamp    time.Time      `gorm:"column:timestamp;not null;index" json:"timestamp" example:"2026-03-20T10:15:00Z"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
