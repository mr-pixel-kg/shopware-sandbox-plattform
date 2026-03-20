package dto

type ImagePayload struct {
	Name         string  `json:"name" validate:"required" example:"dockware/dev"`
	Tag          string  `json:"tag" validate:"required" example:"6.6.9.0"`
	Title        *string `json:"title" example:"Shopware 6.6 Demo"`
	Description  *string `json:"description" example:"Base image for internal sales demos."`
	ThumbnailURL *string `json:"thumbnailUrl" example:"https://cdn.example.com/images/shopware-demo.png"`
	IsPublic     bool    `json:"isPublic" example:"true"`
}

type CreateImageRequest struct {
	ImagePayload
}
