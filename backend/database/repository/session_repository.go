package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/models"
	"log"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ipAddress string, userAgent string, username *string, sandboxId string) error {
	session := &models.Session{
		IpAddress: ipAddress,
		UserAgent: userAgent,
		Username:  username,
		SandboxID: sandboxId,
	}
	query := `INSERT INTO sessions (ip_address, user_agent, username, sandbox_id) VALUES (:ip_address, :user_agent, :username, :sandbox_id)`
	_, err := r.db.NamedExec(query, session)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		return err
	}
	return nil
}

func (r *SessionRepository) GetSessionsForIp(ipAddress string) ([]models.Session, error) {
	var sessions []models.Session
	query := `SELECT * FROM sessions WHERE ip_address = $1`
	err := r.db.Select(&sessions, query, ipAddress)
	if err != nil {
		log.Printf("Error getting all images: %v", err)
		return nil, err
	}
	return sessions, nil
}

func (r *SessionRepository) Remove(sandboxId string) error {
	query := `DELETE FROM sessions WHERE sandbox_id = $1`
	_, err := r.db.Exec(query, sandboxId)
	if err != nil {
		log.Printf("Error deleting image: %v", err)
		return err
	}
	return nil
}
