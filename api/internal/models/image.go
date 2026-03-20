package models

import "github.com/google/uuid"

type Image struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Name            string     `gorm:"size:255;not null" json:"name" example:"dockware/dev"`
	Tag             string     `gorm:"size:255;not null" json:"tag" example:"6.6.9.0"`
	Title           *string    `gorm:"size:255" json:"title,omitempty" example:"Shopware Demo Image"`
	Description     *string    `gorm:"type:text" json:"description,omitempty" example:"Prepared image for sales demos and internal QA."`
	ThumbnailURL    *string    `gorm:"size:1024" json:"thumbnailUrl,omitempty" example:"https://cdn.example.com/images/shopware-demo.png"`
	IsPublic        bool       `gorm:"not null;default:false" json:"isPublic" example:"true"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid" json:"createdByUserId,omitempty" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	BaseModel
}

func (Image) TableName() string {
	return "images"
}

func (i Image) FullName() string {
	return i.Name + ":" + i.Tag
}
