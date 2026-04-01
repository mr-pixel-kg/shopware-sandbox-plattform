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
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestAuditLogRepositoryListOrdersPaginatedResultsAndCountsTotal(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	repo := NewAuditLogRepository(db)
	user := createAuditTestUser(t, db)
	now := time.Now().UTC()

	older := createAuditLogFixture(t, repo, auditLogFixture{
		UserID:    &user.ID,
		Action:    "sandbox.created",
		Timestamp: now.Add(-2 * time.Hour),
	})
	middle := createAuditLogFixture(t, repo, auditLogFixture{
		UserID:    &user.ID,
		Action:    "sandbox.updated",
		Timestamp: now.Add(-1 * time.Hour),
	})
	newest := createAuditLogFixture(t, repo, auditLogFixture{
		UserID:    &user.ID,
		Action:    "sandbox.deleted",
		Timestamp: now,
	})

	logs, total, err := repo.List(AuditLogListOptions{
		Limit:  2,
		Offset: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, logs, 2)
	assert.Equal(t, middle.ID, logs[0].ID)
	assert.Equal(t, older.ID, logs[1].ID)
	assert.Equal(t, newest.UserID, logs[0].UserID)
	require.NotNil(t, logs[0].User)
	assert.Equal(t, user.Email, logs[0].User.Email)
}

func TestAuditLogRepositoryListAppliesAllFilters(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	repo := NewAuditLogRepository(db)
	user := createAuditTestUser(t, db)
	otherUser := createAuditTestUser(t, db)
	resourceID := uuid.New()
	clientToken := uuid.New()
	now := time.Now().UTC()
	resourceType := "sandbox"

	expected := createAuditLogFixture(t, repo, auditLogFixture{
		UserID:       &user.ID,
		Action:       "sandbox.deleted",
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		ClientToken:  &clientToken,
		Timestamp:    now.Add(-30 * time.Minute),
	})
	createAuditLogFixture(t, repo, auditLogFixture{
		UserID:       &user.ID,
		Action:       "sandbox.deleted",
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Timestamp:    now.Add(-31 * time.Minute),
	})
	createAuditLogFixture(t, repo, auditLogFixture{
		UserID:       &otherUser.ID,
		Action:       "sandbox.deleted",
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		ClientToken:  &clientToken,
		Timestamp:    now.Add(-29 * time.Minute),
	})

	from := now.Add(-45 * time.Minute)
	to := now.Add(-15 * time.Minute)
	action := "sandbox.deleted"
	logs, total, err := repo.List(AuditLogListOptions{
		Limit:        50,
		UserID:       &user.ID,
		Action:       &action,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		ClientToken:  &clientToken,
		From:         &from,
		To:           &to,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
	assert.Equal(t, expected.ID, logs[0].ID)
}

type auditLogFixture struct {
	UserID       *uuid.UUID
	Action       string
	ResourceType *string
	ResourceID   *uuid.UUID
	ClientToken  *uuid.UUID
	Timestamp    time.Time
}

func createAuditLogFixture(t *testing.T, repo *AuditLogRepository, fixture auditLogFixture) models.AuditLog {
	t.Helper()

	logEntry := models.AuditLog{
		ID:           uuid.New(),
		UserID:       fixture.UserID,
		Action:       fixture.Action,
		ResourceType: fixture.ResourceType,
		ResourceID:   fixture.ResourceID,
		ClientToken:  fixture.ClientToken,
		Details:      datatypes.JSON([]byte(`{}`)),
		Timestamp:    fixture.Timestamp,
	}
	require.NoError(t, repo.Create(&logEntry))
	return logEntry
}

func createAuditTestUser(t *testing.T, db *gorm.DB) *models.User {
	t.Helper()

	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("audit-user-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "hashed-password",
		Role:         models.RoleUser,
	}
	require.NoError(t, NewUserRepository(db).Create(user))
	return user
}
