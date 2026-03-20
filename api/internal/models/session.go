package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"6c933bc2-ef77-4870-aeb0-2a14f932578f"`
	UserID      *uuid.UUID `gorm:"type:uuid;index" json:"userId,omitempty" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	SessionType string     `gorm:"size:32;not null;index" json:"sessionType" example:"user"`
	TokenID     string     `gorm:"size:255;not null;uniqueIndex" json:"tokenId" example:"01JQ0DY6N8BKA8J5HV0Y5EF7NQ"`
	ExpiresAt   time.Time  `gorm:"not null;index" json:"expiresAt" example:"2026-03-21T10:30:00Z"`
	BaseModel
}

func (Session) TableName() string {
	return "sessions"
}
