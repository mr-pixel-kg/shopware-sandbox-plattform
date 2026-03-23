package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"gorm.io/gorm"
)

type SandboxRepository struct {
	db *gorm.DB
}

func NewSandboxRepository(db *gorm.DB) *SandboxRepository {
	return &SandboxRepository{db: db}
}

func (r *SandboxRepository) Create(sandbox *models.Sandbox) error {
	return r.db.Create(sandbox).Error
}

func (r *SandboxRepository) Update(sandbox *models.Sandbox) error {
	return r.db.Save(sandbox).Error
}

func (r *SandboxRepository) FindByID(id uuid.UUID) (*models.Sandbox, error) {
	var sandbox models.Sandbox
	if err := r.db.First(&sandbox, "id = ?", id).Error; err != nil {
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
	var sandboxes []models.Sandbox
	err := r.db.Where("status IN ?", []models.SandboxStatus{
		models.SandboxStatusStarting,
		models.SandboxStatusRunning,
	}).Order("created_at desc").Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListActiveByUser(userID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.db.
		Where("created_by_user_id = ?", userID).
		Where("status IN ?", []models.SandboxStatus{models.SandboxStatusStarting, models.SandboxStatusRunning}).
		Order("created_at desc").
		Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListAllByUser(userID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.db.
		Where("created_by_user_id = ?", userID).
		Order("created_at desc").
		Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListActiveByGuestSession(sessionID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.db.
		Where("guest_session_id = ?", sessionID).
		Where("status IN ?", []models.SandboxStatus{models.SandboxStatusStarting, models.SandboxStatusRunning}).
		Order("created_at desc").
		Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) ListAllByGuestSession(sessionID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.db.
		Where("guest_session_id = ?", sessionID).
		Order("created_at desc").
		Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) CountActiveByUser(userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Sandbox{}).
		Where("created_by_user_id = ?", userID).
		Where("status IN ?", []models.SandboxStatus{models.SandboxStatusStarting, models.SandboxStatusRunning}).
		Count(&count).Error
	return count, err
}

func (r *SandboxRepository) CountActiveByIP(ip string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Sandbox{}).
		Where("client_ip = ?", ip).
		Where("created_by_user_id IS NULL").
		Where("status IN ?", []models.SandboxStatus{models.SandboxStatusStarting, models.SandboxStatusRunning}).
		Count(&count).Error
	return count, err
}

func (r *SandboxRepository) CountActiveTotal() (int64, error) {
	var count int64
	err := r.db.Model(&models.Sandbox{}).
		Where("status IN ?", []models.SandboxStatus{models.SandboxStatusStarting, models.SandboxStatusRunning}).
		Count(&count).Error
	return count, err
}

func (r *SandboxRepository) ListByImageID(imageID uuid.UUID) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.db.Where("image_id = ?", imageID).Find(&sandboxes).Error
	return sandboxes, err
}

func (r *SandboxRepository) DeleteByID(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&models.Sandbox{}, "id = ?", id).Error
}

func (r *SandboxRepository) FindExpired(now time.Time) ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	err := r.db.
		Where("expires_at IS NOT NULL").
		Where("expires_at <= ?", now).
		Where("status IN ?", []models.SandboxStatus{models.SandboxStatusStarting, models.SandboxStatusRunning}).
		Find(&sandboxes).Error
	return sandboxes, err
}
