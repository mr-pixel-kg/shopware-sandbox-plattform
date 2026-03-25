package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"gorm.io/gorm"
)

type UserService struct {
	users     *repositories.UserRepository
	passwords *PasswordService
}

func NewUserService(users *repositories.UserRepository, passwords *PasswordService) *UserService {
	return &UserService{
		users:     users,
		passwords: passwords,
	}
}

func (s *UserService) List() ([]models.User, error) {
	return s.users.List()
}

func (s *UserService) Get(id uuid.UUID) (*models.User, error) {
	user, err := s.users.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("USER_NOT_FOUND", "User not found").WithCause(err)
		}
		return nil, apperror.Internal("USER_LOOKUP_FAILED", "Could not load user").WithCause(err)
	}

	return user, nil
}

func (s *UserService) Create(email, role string, password *string) (*models.User, error) {
	email = strings.TrimSpace(email)
	role = strings.TrimSpace(role)
	if err := validateManagedUser(email, role, password); err != nil {
		return nil, err
	}

	user := &models.User{
		ID:    uuid.New(),
		Email: email,
		Role:  role,
	}

	if password != nil {
		passwordHash, err := s.passwords.Hash(*password)
		if err != nil {
			return nil, apperror.Internal("PASSWORD_HASH_FAILED", "Could not store password").WithCause(err)
		}
		user.PasswordHash = passwordHash
	}

	if err := s.users.Create(user); err != nil {
		return nil, apperror.Conflict("EMAIL_EXISTS", "Email already exists").WithCause(err)
	}

	return user, nil
}

func (s *UserService) Update(id uuid.UUID, email, role string, password *string) (*models.User, error) {
	email = strings.TrimSpace(email)
	role = strings.TrimSpace(role)
	if err := validateManagedUser(email, role, password); err != nil {
		return nil, err
	}

	user, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	user.Email = email
	user.Role = role

	if password != nil {
		passwordHash, hashErr := s.passwords.Hash(*password)
		if hashErr != nil {
			return nil, apperror.Internal("PASSWORD_HASH_FAILED", "Could not store password").WithCause(hashErr)
		}
		user.PasswordHash = passwordHash
	}

	if err := s.users.Update(user); err != nil {
		return nil, apperror.Conflict("EMAIL_EXISTS", "Email already exists").WithCause(err)
	}

	return user, nil
}

func (s *UserService) Delete(id uuid.UUID) error {
	user, err := s.Get(id)
	if err != nil {
		return err
	}

	if err := s.users.Delete(user.ID); err != nil {
		return apperror.Internal("USER_DELETE_FAILED", "Could not delete user").WithCause(err)
	}

	return nil
}

func validateManagedUser(email, role string, password *string) *apperror.AppError {
	if email == "" {
		return apperror.BadRequest("EMAIL_REQUIRED", "Email is required")
	}

	if role != models.RoleAdmin && role != models.RoleUser {
		return apperror.BadRequest("INVALID_ROLE", "Role must be 'admin' or 'user'")
	}

	if password != nil && strings.TrimSpace(*password) == "" {
		return apperror.BadRequest("INVALID_PASSWORD", "Password must not be empty when provided")
	}

	return nil
}
