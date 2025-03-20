package services

import (
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/models"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/repository"
	"log"
)

const MAX_SESSIONS_PER_IP = 3

type GuardService struct {
	SessionRepository *repository.SessionRepository
}

func NewGuardService(sessionRepository *repository.SessionRepository) *GuardService {
	return &GuardService{
		SessionRepository: sessionRepository,
	}
}

// Gets all sandbox sessions for a given IP address
func (s *GuardService) GetSessions(ipAddress string) []models.Session {
	sessions, err := s.SessionRepository.GetSessionsForIp(ipAddress)
	if err != nil {
		log.Printf("Error getting sessions for IP: %v", err)
		return make([]models.Session, 0)
	}
	return sessions
}

// Records a new session after new sandbox is created
func (s *GuardService) RegisterSession(ipAddress string, userAgent string, username *string, sandboxId string) error {
	err := s.SessionRepository.Create(ipAddress, userAgent, username, sandboxId)
	if err != nil {
		return err
	}
	return nil
}

// Removes a sandbox session when the sandbox is deleted
func (s *GuardService) UnregisterSession(sandboxId string) error {
	err := s.SessionRepository.Remove(sandboxId)
	if err != nil {
		return err
	}
	return nil
}

// Checks if the current session (based on IP address) has exceeded the limit of concurrent sandboxes
// Returns true if the limit is exceeded
// Returns false if the limit is not exceeded
func (s *GuardService) IsNewSessionAllowed(ipAddress string) bool {
	sessions := s.GetSessions(ipAddress)
	if len(sessions) < MAX_SESSIONS_PER_IP {
		return true
	}
	return false
}

// Checks if the current IP address has exceeded the limit of concurrent sandboxes
// Returns true and records a new session if the limit is not exceeded
// Returns false if the limit is exceeded for that IP address
func (s *GuardService) CheckAndRegisterSession(ipAddress string, userAgent string, username *string, sandboxId string) (bool, error) {
	if s.IsNewSessionAllowed(ipAddress) {
		err := s.RegisterSession(ipAddress, userAgent, username, sandboxId)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}
