package dto

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"testmail@gmail.com"`
	Password string `json:"password" validate:"required,min=8" example:"test123123"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"testmail@gmail.com"`
	Password string `json:"password" validate:"required" example:"test123123"`
}

type LoginResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  UserResponse `json:"user"`
}
