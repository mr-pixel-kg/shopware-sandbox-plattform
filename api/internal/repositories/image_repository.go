package repositories

import (
	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"gorm.io/gorm"
)

type ImageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) withOwner(db *gorm.DB) *gorm.DB {
	return db.Preload("Owner", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "email")
	})
}

func (r *ImageRepository) ListPublic() ([]models.Image, error) {
	var images []models.Image
	err := r.withOwner(r.db).
		Where("is_public = ? AND status = ?", true, models.ImageStatusReady).
		Order("created_at desc").
		Find(&images).Error
	return images, err
}

func (r *ImageRepository) ListAll() ([]models.Image, error) {
	var images []models.Image
	err := r.withOwner(r.db).Order("created_at desc").Find(&images).Error
	return images, err
}

func (r *ImageRepository) FindByID(id uuid.UUID) (*models.Image, error) {
	var image models.Image
	if err := r.withOwner(r.db).First(&image, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *ImageRepository) Create(image *models.Image) error {
	return r.db.Create(image).Error
}

func (r *ImageRepository) Update(image *models.Image) error {
	return r.db.Save(image).Error
}

func (r *ImageRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Image{}, "id = ?", id).Error
}

func (r *ImageRepository) ListByStatus(status string) ([]models.Image, error) {
	var images []models.Image
	err := r.db.Where("status = ?", status).Order("created_at desc").Find(&images).Error
	return images, err
}

func (r *ImageRepository) ListByStatuses(statuses []string) ([]models.Image, error) {
	var images []models.Image
	err := r.db.Where("status IN ?", statuses).Order("created_at desc").Find(&images).Error
	return images, err
}

func (r *ImageRepository) UpdateStatus(id uuid.UUID, status string, errMsg *string) error {
	return r.db.Model(&models.Image{}).Where("id = ?", id).Updates(map[string]any{
		"status": status,
		"error":  errMsg,
	}).Error
}

func (r *ImageRepository) ResolveFullName(id uuid.UUID) string {
	var img models.Image
	if err := r.db.First(&img, "id = ?", id).Error; err != nil {
		return ""
	}
	return img.FullName()
}

func (r *ImageRepository) ResolveRegistryName(id uuid.UUID) string {
	var img models.Image
	if err := r.db.First(&img, "id = ?", id).Error; err != nil {
		return ""
	}
	return img.RegistryName()
}
