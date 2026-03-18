package models

import "github.com/google/uuid"

type Image struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Name            string     `gorm:"size:255;not null" json:"name"`
	Tag             string     `gorm:"size:255;not null" json:"tag"`
	Title           *string    `gorm:"size:255" json:"title,omitempty"`
	Description     *string    `gorm:"type:text" json:"description,omitempty"`
	ThumbnailURL    *string    `gorm:"size:1024" json:"thumbnailUrl,omitempty"`
	IsPublic        bool       `gorm:"not null;default:false" json:"isPublic"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid" json:"createdByUserId,omitempty"`
	BaseModel
}

func (Image) TableName() string {
	return "images"
}

func (i Image) FullName() string {
	return i.Name + ":" + i.Tag
}
