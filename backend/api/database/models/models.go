package models

import "time"

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
