package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"gorm.io/gorm"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users     *repositories.UserRepository
	sessions  *repositories.SessionRepository
	passwords *PasswordService
	tokens    *TokenService
}

func NewAuthService(
	users *repositories.UserRepository,
	sessions *repositories.SessionRepository,
	passwords *PasswordService,
	tokens *TokenService,
) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		passwords: passwords,
		tokens:    tokens,
	}
}

func (s *AuthService) Register(email, password string) (*models.User, error) {
	passwordHash, err := s.passwords.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(email, password string) (string, *models.User, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := s.passwords.Verify(user.PasswordHash, password); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, tokenID, expiresAt, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", nil, err
	}

	// A persisted session lets us invalidate tokens server-side on logout even
	// though the API uses JWTs for request authentication.
	if err := s.sessions.Create(&models.Session{
		ID:          uuid.New(),
		UserID:      &user.ID,
		SessionType: "user",
		TokenID:     tokenID,
		ExpiresAt:   expiresAt,
	}); err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) Authenticate(tokenValue string) (*models.User, string, error) {
	claims, err := s.tokens.Parse(tokenValue)
	if err != nil {
		return nil, "", err
	}

	// JWT validation alone is not enough because sessions may have been revoked.
	valid, err := s.sessions.ExistsActiveToken(claims.TokenID, time.Now().UTC())
	if err != nil {
		return nil, "", err
	}
	if !valid {
		return nil, "", ErrInvalidCredentials
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, "", err
	}

	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, "", err
	}

	return user, claims.TokenID, nil
}

func (s *AuthService) Logout(tokenID string) error {
	return s.sessions.DeleteByTokenID(tokenID)
}
