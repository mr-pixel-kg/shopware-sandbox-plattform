package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/models"
	"log"
)

type AuditLogRepository struct {
	db *sqlx.DB
}

func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(auditLogEntry *models.AuditLogEntry) error {
	query := `INSERT INTO audit_log (ip_address, user_agent, username, action, details) VALUES (:ip_address, :user_agent, :username, :action, :details)`
	_, err := r.db.NamedExec(query, auditLogEntry)
	if err != nil {
		log.Printf("Error creating image: %v", err)
		return err
	}
	return nil
}

func (r *AuditLogRepository) GetAll() ([]models.AuditLogEntry, error) {
	var entries []models.AuditLogEntry
	query := `SELECT * FROM audit_log`
	err := r.db.Select(&entries, query)
	if err != nil {
		log.Printf("Error getting all audit log entries: %v", err)
		return nil, err
	}
	return entries, nil
}
