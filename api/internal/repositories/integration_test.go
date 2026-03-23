//go:build integration

package repositories

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateAndFind(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	repo := NewUserRepository(db)
	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hashed-password",
	}

	require.NoError(t, repo.Create(user))

	byEmail, err := repo.FindByEmail(user.Email)
	require.NoError(t, err)
	assert.Equal(t, user.ID, byEmail.ID)

	byID, err := repo.FindByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, byID.Email)
}

func TestSessionRepository_ExistsActiveTokenAndDelete(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	userRepo := NewUserRepository(db)
	sessionRepo := NewSessionRepository(db)

	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("session-user-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hashed-password",
	}
	require.NoError(t, userRepo.Create(user))

	activeTokenID := uuid.NewString()
	expiredTokenID := uuid.NewString()
	now := time.Now().UTC()

	require.NoError(t, sessionRepo.Create(&models.Session{
		ID:          uuid.New(),
		UserID:      &user.ID,
		SessionType: "user",
		TokenID:     activeTokenID,
		ExpiresAt:   now.Add(30 * time.Minute),
	}))

	require.NoError(t, sessionRepo.Create(&models.Session{
		ID:          uuid.New(),
		UserID:      &user.ID,
		SessionType: "user",
		TokenID:     expiredTokenID,
		ExpiresAt:   now.Add(-30 * time.Minute),
	}))

	active, err := sessionRepo.ExistsActiveToken(activeTokenID, now)
	require.NoError(t, err)
	assert.True(t, active)

	expired, err := sessionRepo.ExistsActiveToken(expiredTokenID, now)
	require.NoError(t, err)
	assert.False(t, expired)

	require.NoError(t, sessionRepo.DeleteByTokenID(activeTokenID))

	active, err = sessionRepo.ExistsActiveToken(activeTokenID, now)
	require.NoError(t, err)
	assert.False(t, active)
}
