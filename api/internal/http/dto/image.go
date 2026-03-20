package dto

type ImagePayload struct {
	Name         string  `json:"name" validate:"required" example:"dockware/dev"`
	Tag          string  `json:"tag" validate:"required" example:"6.6.9.0"`
	Title        *string `json:"title" example:"Shopware 6.6 Demo"`
	Description  *string `json:"description" example:"Base image for internal sales demos."`
	IsPublic     bool    `json:"isPublic" example:"true"`
}

type CreateImageRequest struct {
	ImagePayload
}

type UpdateImageRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	IsPublic    bool    `json:"isPublic"`
}
