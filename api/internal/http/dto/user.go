package dto

type CreateUserRequest struct {
	Email    string  `json:"email" example:"jane.doe@example.com"`
	Role     string  `json:"role" example:"user"`
	Password *string `json:"password,omitempty" example:"Sup3rS3cret!"`
}

type UpdateUserRequest struct {
	Email    string  `json:"email" example:"jane.doe@example.com"`
	Role     string  `json:"role" example:"admin"`
	Password *string `json:"password,omitempty" example:"N3wSup3rS3cret!"`
}
