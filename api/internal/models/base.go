package models

import "time"

type BaseModel struct {
	CreatedAt time.Time  `gorm:"not null" json:"createdAt" example:"2026-03-20T10:15:00Z"`
	UpdatedAt time.Time  `gorm:"not null" json:"updatedAt" example:"2026-03-20T10:20:00Z"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty" swaggerignore:"true"`
}
