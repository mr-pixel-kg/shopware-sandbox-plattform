package repository

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/api/database/models"
	"log"
)

type SandboxRepository struct {
	db *sqlx.DB
}

func NewSandboxRepository(db *sqlx.DB) *SandboxRepository {
	return &SandboxRepository{db: db}
}

func (r *SandboxRepository) Create(sandbox *models.Sandbox) (*models.Sandbox, error) {
	// Convert timestamps to UTC
	sandbox.CreatedAt = sandbox.CreatedAt.UTC()
	if sandbox.DestroyAt != nil {
		*sandbox.DestroyAt = sandbox.DestroyAt.UTC()
	}

	query := `INSERT INTO sandboxes (id, container_id, container_name, image_id, url, created_at, destroy_at) VALUES (:id, :container_id, :container_name, :image_id, :url, :created_at, :destroy_at)`
	_, err := r.db.NamedExec(query, sandbox)
	if err != nil {
		log.Printf("Error creating sandbox: %v", err)
		return nil, err
	}
	return sandbox, nil
}

func (r *SandboxRepository) GetByID(id string) (*models.Sandbox, error) {
	var sandbox models.Sandbox
	query := `SELECT * FROM sandboxes WHERE id = $1`
	err := r.db.Get(&sandbox, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("Error getting sandbox by ID: %v", err)
		return nil, err
	}
	return &sandbox, nil
}

func (r *SandboxRepository) GetAll() ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox
	query := `SELECT * FROM sandboxes`
	err := r.db.Select(&sandboxes, query)
	if err != nil {
		log.Printf("Error getting all sandboxes: %v", err)
		return nil, err
	}
	return sandboxes, nil
}

func (r *SandboxRepository) Update(sandbox *models.Sandbox) error {
	query := `UPDATE sandboxes SET container_id = :container_id, container_name = :container_name, image_id = :image_id, url = :url, created_at = :created_at, destroy_at = :destroy_at WHERE id = :id`
	_, err := r.db.NamedExec(query, sandbox)
	if err != nil {
		log.Printf("Error updating sandbox: %v", err)
		return err
	}
	return nil
}

func (r *SandboxRepository) Delete(id string) error {
	query := `DELETE FROM sandboxes WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting sandbox: %v", err)
		return err
	}
	return nil
}

func (r *SandboxRepository) GetExpiredContainers() ([]models.Sandbox, error) {
	var sandboxes []models.Sandbox

	// Prepare query dialect
	sqlNowStatement := "NOW()"
	if r.db.DriverName() == "sqlite3" {
		sqlNowStatement = "strftime('%Y-%m-%d %H:%M:%S', 'now')"
	}

	// Execute query
	query := fmt.Sprintf(`SELECT * FROM sandboxes WHERE destroy_at IS NOT NULL AND destroy_at < %s`, sqlNowStatement)
	err := r.db.Select(&sandboxes, query)
	if err != nil {
		log.Printf("Error getting expired containers: %v", err)
		return nil, err
	}
	return sandboxes, nil
}
