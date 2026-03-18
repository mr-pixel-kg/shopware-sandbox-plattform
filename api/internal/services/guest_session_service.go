package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
)

var ErrInvalidGuestSession = errors.New("invalid guest session")

type GuestSessionService struct {
	sessions *repositories.SessionRepository
	tokens   *TokenService
}

func NewGuestSessionService(sessions *repositories.SessionRepository, tokens *TokenService) *GuestSessionService {
	return &GuestSessionService{
		sessions: sessions,
		tokens:   tokens,
	}
}

func (s *GuestSessionService) Ensure(tokenValue string) (string, uuid.UUID, error) {
	if tokenValue != "" {
		// Reuse the existing guest identity whenever the cookie is still valid so
		// anonymous users keep seeing the same sandbox ownership.
		sessionID, _, err := s.Validate(tokenValue)
		if err == nil {
			return tokenValue, sessionID, nil
		}
	}

	sessionID := uuid.New()
	token, tokenID, expiresAt, err := s.tokens.GenerateGuest(sessionID)
	if err != nil {
		return "", uuid.Nil, err
	}

	if err := s.sessions.Create(&models.Session{
		ID:          sessionID,
		SessionType: "guest",
		TokenID:     tokenID,
		ExpiresAt:   expiresAt,
	}); err != nil {
		return "", uuid.Nil, err
	}

	return token, sessionID, nil
}

func (s *GuestSessionService) Validate(tokenValue string) (uuid.UUID, string, error) {
	claims, err := s.tokens.Parse(tokenValue)
	if err != nil {
		return uuid.Nil, "", err
	}
	if claims.SessionType != "guest" {
		return uuid.Nil, "", ErrInvalidGuestSession
	}

	// Guest tokens are stored in the same session table so cleanup and expiry
	// behaviour stay consistent with authenticated employee sessions.
	valid, err := s.sessions.ExistsActiveToken(claims.TokenID, time.Now().UTC())
	if err != nil {
		return uuid.Nil, "", err
	}
	if !valid {
		return uuid.Nil, "", ErrInvalidGuestSession
	}

	sessionID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", err
	}
	return sessionID, claims.TokenID, nil
}
