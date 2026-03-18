package dto

type ImagePayload struct {
	Name         string  `json:"name"`
	Tag          string  `json:"tag"`
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	ThumbnailURL *string `json:"thumbnailUrl"`
	IsPublic     bool    `json:"isPublic"`
}

type CreateImageRequest struct {
	ImagePayload
}
