package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"gorm.io/gorm"
)

var activeStatuses = []models.SandboxStatus{
	models.SandboxStatusStarting,
	models.SandboxStatusRunning,
	models.SandboxStatusPaused,
	models.SandboxStatusStopping,
}

type SandboxRepository struct {
	db *gorm.DB
}

func NewSandboxRepository(db *gorm.DB) *SandboxRepository {
	return &SandboxRepository{db: db}
}

func (r *SandboxRepository) withOwner(db *gorm.DB) *gorm.DB {
	return db.Preload("Owner", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "email")
	})
}

func (r *SandboxRepository) whereActive(db *gorm.DB) *gorm.DB {
	return db.Where("status IN ?", activeStatuses)
}

func (r *SandboxRepository) Create(sandbox *models.Sandbox) error {
	return r.db.Create(sandbox).Error
}

func (r *SandboxRepository) Update(sandbox *models.Sandbox) error {
	return r.db.Save(sandbox).Error
}

func (r *SandboxRepository) FindByID(id uuid.UUID) (*models.Sandbox, error) {
	var sandbox models.Sandbox
	if err := r.withOwner(r.db).First(&sandbox, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sandbox, nil
}

func (r *SandboxRepository) FindByContainerID(containerID string) (*models.Sandbox, error) {
	var sandbox models.Sandbox
	if err := r.db.First(&sandbox, "container_id = ?", containerID).Error; err != nil {
		return nil, err
	}
	return &sandbox, nil
}

func (r *SandboxRepository) ListAllActive() ([]models.Sandbox, error) {
	return r.ListByStatuses(activeStatuses)
}

func (r *SandboxRepository) ListAllByUser(userID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.withOwner(r.db).
		Where("owner_id = ?", userID).
		Order("created_at desc").
		Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListAllByClientID(clientID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.withOwner(r.db).
		Where("client_id = ?", clientID).
		Order("created_at desc").
		Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) CountActiveByUser(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.whereActive(r.db.Model(&models.Sandbox{})).
		Where("owner_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *SandboxRepository) CountActiveByIP(ip string) (int64, error) {
	var count int64
	err := r.whereActive(r.db.Model(&models.Sandbox{})).
		Where("client_ip = ?", ip).
		Where("owner_id IS NULL").
		Count(&count).Error
	return count, err
}

func (r *SandboxRepository) CountActiveTotal() (int64, error) {
	var count int64
	err := r.whereActive(r.db.Model(&models.Sandbox{})).
		Count(&count).Error
	return count, err
}

func (r *SandboxRepository) ListByImageID(imageID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.db.Where("image_id = ?", imageID).Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListByStatuses(statuses []models.SandboxStatus) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.withOwner(r.db).
		Where("status IN ?", statuses).
		Order("created_at desc").
		Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListAll() ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.withOwner(r.db).Order("created_at desc").Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListAllPaginated(limit, offset int) ([]models.Sandbox, int64, error) {
	var sandboxes []models.Sandbox
	var total int64
	query := r.withOwner(r.db).Order("created_at desc")
	if err := query.Model(&models.Sandbox{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Limit(limit).Offset(offset).Find(&sandboxes).Error
	return sandboxes, total, err
}

func (r *SandboxRepository) ListAllByUserPaginated(userID uuid.UUID, limit, offset int) ([]models.Sandbox, int64, error) {
	var sandboxes []models.Sandbox
	var total int64
	query := r.withOwner(r.db).Where("owner_id = ?", userID).Order("created_at desc")
	if err := query.Model(&models.Sandbox{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Limit(limit).Offset(offset).Find(&sandboxes).Error
	return sandboxes, total, err
}

func (r *SandboxRepository) ListAllByOwnerPaginated(userID, clientID uuid.UUID, limit, offset int) ([]models.Sandbox, int64, error) {
	var sandboxes []models.Sandbox
	var total int64
	query := r.withOwner(r.db).Where("owner_id = ? OR client_id = ?", userID, clientID).Order("created_at desc")
	if err := query.Model(&models.Sandbox{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Limit(limit).Offset(offset).Find(&sandboxes).Error
	return sandboxes, total, err
}

func (r *SandboxRepository) ListAllByClientIDPaginated(clientID uuid.UUID, limit, offset int) ([]models.Sandbox, int64, error) {
	var sandboxes []models.Sandbox
	var total int64
	query := r.withOwner(r.db).Where("client_id = ?", clientID).Order("created_at desc")
	if err := query.Model(&models.Sandbox{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Limit(limit).Offset(offset).Find(&sandboxes).Error
	return sandboxes, total, err
}

func (r *SandboxRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&models.Sandbox{}, "id = ?", id).Error
}

func (r *SandboxRepository) FindExpired(now time.Time) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.whereActive(r.db).
		Where("expires_at IS NOT NULL").
		Where("expires_at <= ?", now).
		Find(&sandboxes).Error
	return sandboxes, err
}
