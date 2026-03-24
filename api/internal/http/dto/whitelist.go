package dto

type AddWhitelistRequest struct {
	Email string `json:"email" validate:"required,email" example:"jane.doe@example.com"`
	Role  string `json:"role" validate:"required,oneof=admin user" example:"user"`
}
