package dto

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GravatarURL(email string, size int) string {
	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("https://gravatar.com/avatar/%x?s=%d&d=retro", hash, size)
}

type UserSummary struct {
	ID        uuid.UUID `json:"id" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	Email     string    `json:"email" example:"testmail@gmail.com"`
	AvatarURL string    `json:"avatarUrl" example:"https://gravatar.com/avatar/abc123?s=80&d=retro"`
}

type UserListResponse struct {
	Data []UserResponse `json:"data"`
	Meta PaginatedMeta  `json:"meta"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	Email     string    `json:"email" example:"testmail@gmail.com"`
	AvatarURL string    `json:"avatarUrl" example:"https://gravatar.com/avatar/abc123?s=80&d=retro"`
	Role      string    `json:"role" example:"user"`
	IsPending bool      `json:"isPending" example:"false"`
	CreatedAt time.Time `json:"createdAt" example:"2026-03-20T10:15:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2026-03-20T10:20:00Z"`
}

type CreateUserRequest struct {
	Email    string  `json:"email" validate:"required,email" example:"testmail@gmail.com"`
	Role     string  `json:"role" validate:"required,oneof=admin user" example:"user"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8" example:"test123123"`
}

type UpdateUserRequest struct {
	Email    string  `json:"email" validate:"required,email" example:"testmail@gmail.com"`
	Role     string  `json:"role" validate:"required,oneof=admin user" example:"admin"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8" example:"test123123"`
}
