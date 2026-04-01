package dto

type CreateUserRequest struct {
	Email    string  `json:"email" validate:"required,email" example:"jane.doe@example.com"`
	Role     string  `json:"role" validate:"required,oneof=admin user" example:"user"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8" example:"Sup3rS3cret!"`
}

type UpdateUserRequest struct {
	Email    string  `json:"email" validate:"required,email" example:"jane.doe@example.com"`
	Role     string  `json:"role" validate:"required,oneof=admin user" example:"admin"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8" example:"N3wSup3rS3cret!"`
}
