package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID        int       `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"password"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Image struct {
	ID        string    `db:"id" json:"id"`
	ImageName string    `db:"image_name" json:"image_name"`
	ImageTag  string    `db:"image_tag" json:"image_tag"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Sandbox struct {
	ID            string     `db:"id" json:"id"`
	ContainerID   string     `db:"container_id" json:"container_id"`
	ContainerName string     `db:"container_name" json:"container_name"`
	ImageID       string     `db:"image_id" json:"image_id"`
	URL           string     `db:"url" json:"url"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	DestroyAt     *time.Time `db:"destroy_at" json:"destroy_at"`
}

type AuditAction string

const (
	USER_LOGIN            AuditAction = "USER_LOGIN"
	USER_LOGIN_FAILED     AuditAction = "USER_LOGIN_FAILED"
	SANDBOX_CREATE        AuditAction = "SANDBOX_CREATE"
	SANDBOX_DELETE        AuditAction = "SANDBOX_DELETE"
	IMAGE_CREATE          AuditAction = "IMAGE_CREATE"
	IMAGE_DELETE          AuditAction = "IMAGE_DELETE"
	CONTAINER_AUTO_REMOVE AuditAction = "CONTAINER_AUTO_REMOVE"
)

type AuditLogEntry struct {
	ID        int             `db:"id" json:"id"`
	Timestamp time.Time       `db:"timestamp" json:"timestamp"`
	IpAddress string          `db:"ip_address" json:"ip_address"`
	UserAgent string          `db:"user_agent" json:"user_agent"`
	Username  *string         `db:"username" json:"username"`
	Action    AuditAction     `db:"action" json:"action"`
	Details   json.RawMessage `db:"details" json:"details"`
}

type Session struct {
	ID        int     `db:"id" json:"id"`
	IpAddress string  `db:"ip_address" json:"ip_address"`
	UserAgent string  `db:"user_agent" json:"user_agent"`
	Username  *string `db:"username" json:"username"`
	SandboxID string  `db:"sandbox_id" json:"sandbox_id"`
}
