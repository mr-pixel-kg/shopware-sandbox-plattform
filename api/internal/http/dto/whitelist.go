package dto

type AddWhitelistRequest struct {
	Email string `json:"email" validate:"required,email" example:"testmail@gmail.com"`
	Role  string `json:"role" validate:"required,oneof=admin user" example:"user"`
}
