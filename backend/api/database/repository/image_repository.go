package repository

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database/models"
	"log"
	"time"
)

type ImageRepository struct {
	db *sqlx.DB
}

func NewImageRepository(db *sqlx.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) Create(image_name, image_tag string) (*models.Image, error) {
	image := &models.Image{
		ID:        uuid.New().String(),
		ImageName: image_name,
		ImageTag:  image_tag,
		CreatedAt: time.Now(),
	}
	query := `INSERT INTO images (id, image_name, image_tag, created_at) VALUES (:id, :image_name, :image_tag, :created_at)`
	_, err := r.db.NamedExec(query, image)
	if err != nil {
		log.Printf("Error creating image: %v", err)
		return nil, err
	}
	return image, nil
}

func (r *ImageRepository) GetByID(id string) (*models.Image, error) {
	var image models.Image
	query := `SELECT * FROM images WHERE id = $1`
	err := r.db.Get(&image, query, id)
	if err != nil {
		log.Printf("Error getting image by ID: %v", err)
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) GetByNameAndTag(name, tag string) (*models.Image, error) {
	var image models.Image
	query := `SELECT * FROM images WHERE image_name = $1 AND image_tag = $2`
	err := r.db.Get(&image, query, name, tag)
	if err != nil {
		log.Printf("Error getting image by name and tag: %v", err)
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) GetAll() ([]models.Image, error) {
	var images []models.Image
	query := `SELECT * FROM images`
	err := r.db.Select(&images, query)
	if err != nil {
		log.Printf("Error getting all images: %v", err)
		return nil, err
	}
	return images, nil
}

func (r *ImageRepository) Update(image *models.Image) error {
	query := `UPDATE images SET image_name = :image_name, image_tag = :image_tag, created_at = :created_at WHERE id = :id`
	_, err := r.db.NamedExec(query, image)
	if err != nil {
		log.Printf("Error updating image: %v", err)
		return err
	}
	return nil
}

func (r *ImageRepository) Delete(id string) error {
	query := `DELETE FROM images WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting image: %v", err)
		return err
	}
	return nil
}

func (r *ImageRepository) DeleteByTagAndName(image_name, image_tag string) error {
	query := `DELETE FROM images WHERE image_name = $1 AND image_tag = $2`
	_, err := r.db.Exec(query, image_name, image_tag)
	if err != nil {
		log.Printf("Error deleting image: %v", err)
		return err
	}
	return nil
}

func (r *ImageRepository) IsAllowed(imageName, imageTag string) bool {
	var image models.Image
	query := `SELECT * FROM images WHERE image_name = $1 AND image_tag = $2`
	err := r.db.Get(&image, query, imageName, imageTag)

	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Fatalln("Failed to check if image is on whitelist", err)
		return false
	}
	return true
}
