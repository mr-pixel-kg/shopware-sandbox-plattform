package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
	"gorm.io/datatypes"
)

type ImagePayload struct {
	Name        string                  `json:"name" validate:"required" example:"dockware/dev"`
	Tag         string                  `json:"tag" validate:"required" example:"6.6.9.0"`
	Title       *string                 `json:"title" example:"Shopware 6.6 Demo"`
	Description *string                 `json:"description" example:"Base image for internal sales demos."`
	IsPublic    bool                    `json:"isPublic" example:"true"`
	Metadata    []registry.MetadataItem `json:"metadata,omitempty"`
}

type CreateImageRequest struct {
	ImagePayload
}

type UpdateImageRequest struct {
	Title       *string                 `json:"title"`
	Description *string                 `json:"description"`
	IsPublic    bool                    `json:"isPublic"`
	Metadata    []registry.MetadataItem `json:"metadata,omitempty"`
}

type ImageResponse struct {
	ID           uuid.UUID      `json:"id" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Name         string         `json:"name" example:"dockware/dev"`
	Tag          string         `json:"tag" example:"6.6.9.0"`
	Title        *string        `json:"title,omitempty" example:"Shopware Demo Image"`
	Description  *string        `json:"description,omitempty" example:"Prepared image for sales demos and internal QA."`
	ThumbnailURL *string        `json:"thumbnailUrl,omitempty" example:"https://cdn.example.com/images/shopware-demo.png"`
	IsPublic     bool           `json:"isPublic" example:"true"`
	Status       string         `json:"status" example:"ready"`
	Error        *string        `json:"error,omitempty" example:"pull access denied"`
	Metadata     datatypes.JSON `json:"metadata,omitempty" swaggertype:"string"`
	RegistryRef  *string        `json:"registryRef,omitempty" example:"dockware/dev"`
	Owner        *UserSummary   `json:"owner,omitempty"`
	CreatedAt    time.Time      `json:"createdAt" example:"2026-03-20T10:15:00Z"`
	UpdatedAt    time.Time      `json:"updatedAt" example:"2026-03-20T10:20:00Z"`
}

type PendingImageResponse struct {
	ID      string  `json:"id" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Name    string  `json:"name" example:"dockware/shopware"`
	Tag     string  `json:"tag" example:"6.6.9.0"`
	Title   *string `json:"title,omitempty" example:"Shopware Demo Image"`
	Percent int     `json:"percent" example:"42"`
	Status  string  `json:"status" example:"pulling"`
}

type ImageProgressEvent struct {
	Percent int    `json:"percent" example:"75"`
	Status  string `json:"status" example:"pulling"`
}
