package models

import "github.com/google/uuid"

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	Email        string    `gorm:"size:255;uniqueIndex;not null" json:"email" example:"jane.doe@example.com"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         string    `gorm:"size:32;not null;default:''" json:"role" example:"user"`
	BaseModel
}

func (User) TableName() string {
	return "users"
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsPending() bool {
	return u.PasswordHash == ""
}
