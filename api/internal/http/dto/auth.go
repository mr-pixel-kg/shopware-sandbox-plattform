package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"jane.doe@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"Sup3rS3cret!"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"jane.doe@example.com"`
	Password string `json:"password" validate:"required" example:"Sup3rS3cret!"`
}
