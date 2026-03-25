package dto

import "github.com/manuel/shopware-testenv-platform/api/internal/registry"

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
