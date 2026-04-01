package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"gorm.io/gorm"
)

type AuditLogRepository struct {
	db *gorm.DB
}

type AuditLogListOptions struct {
	Limit        int
	Offset       int
	UserID       *uuid.UUID
	Action       *string
	ResourceType *string
	ResourceID   *uuid.UUID
	ClientToken  *uuid.UUID
	From         *time.Time
	To           *time.Time
}

func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(entry *models.AuditLog) error {
	return r.db.Create(entry).Error
}

func (r *AuditLogRepository) List(options AuditLogListOptions) ([]models.AuditLog, int64, error) {
	var (
		logs  []models.AuditLog
		total int64
	)

	query := r.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "email")
	})

	query = applyAuditLogFilters(query, options)

	if err := query.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("timestamp desc").
		Limit(options.Limit).
		Offset(options.Offset).
		Find(&logs).Error
	return logs, total, err
}

func applyAuditLogFilters(query *gorm.DB, options AuditLogListOptions) *gorm.DB {
	if options.UserID != nil {
		query = query.Where("user_id = ?", *options.UserID)
	}
	if options.Action != nil {
		query = query.Where("action = ?", *options.Action)
	}
	if options.ResourceType != nil {
		query = query.Where("resource_type = ?", *options.ResourceType)
	}
	if options.ResourceID != nil {
		query = query.Where("resource_id = ?", *options.ResourceID)
	}
	if options.ClientToken != nil {
		query = query.Where("client_token = ?", *options.ClientToken)
	}
	if options.From != nil {
		query = query.Where("timestamp >= ?", *options.From)
	}
	if options.To != nil {
		query = query.Where("timestamp <= ?", *options.To)
	}

	return query
}
