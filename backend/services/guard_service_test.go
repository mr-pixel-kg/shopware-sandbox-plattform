package services

import (
	"github.com/jmoiron/sqlx"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database"
	"github.com/mr-pixel-kg/shopware-sandbox-plattform/database/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setupGuardService(t *testing.T) *GuardService {
	// Prepare in-memory database
	var err error
	database.DB, err = sqlx.Open("sqlite3", ":memory:")
	assert.NoError(t, err, "Creating database should not return an error")
	createTables(t)

	sessionRepository := repository.NewSessionRepository(database.DB)
	guardService := NewGuardService(sessionRepository)
	assert.NoError(t, err, "Creating docker service should not return an error")
	return guardService
}

func createTables(t *testing.T) {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    ip_address VARCHAR(16) NOT NULL,
	    user_agent VARCHAR(255) NOT NULL,
	    username VARCHAR(64) DEFAULT NULL,
		sandbox_id VARCHAR(255) NOT NULL,
		FOREIGN KEY(sandbox_id) REFERENCES sandboxes(id) ON DELETE CASCADE
	);
	`

	_, err := database.DB.Exec(schema)
	assert.NoError(t, err, "Creating database schema should not return an error")
}

func TestGetSessions(t *testing.T) {
	guardService := setupGuardService(t)

	t.Run("GetSessions should be empty", func(t *testing.T) {
		sessions := guardService.GetSessions("127.0.0.1")
		assert.Empty(t, sessions, "GetSessions should be empty")
	})

	t.Run("GetSessions should return one session", func(t *testing.T) {
		guardService.RegisterSession("127.0.0.1", "Test", nil, "sandbox1")

		sessions := guardService.GetSessions("127.0.0.1")
		assert.Len(t, sessions, 1, "GetSessions should have length 1")
	})

	t.Run("GetSessions should be empty after remove", func(t *testing.T) {
		guardService.UnregisterSession("sandbox1")

		sessions := guardService.GetSessions("127.0.0.1")
		assert.Empty(t, sessions, "GetSessions should be empty")
	})

	t.Run("Test more complex case", func(t *testing.T) {
		res, _ := guardService.CheckAndRegisterSession("127.0.0.1", "Test", nil, "sandbox1")
		assert.True(t, res, "CheckAndRegisterSession should return true")
		sessions := guardService.GetSessions("127.0.0.1")
		assert.Len(t, sessions, 1, "GetSessions should have length 1")

		res, _ = guardService.CheckAndRegisterSession("127.0.0.1", "Test", nil, "sandbox2")
		assert.True(t, res, "CheckAndRegisterSession should return true")
		sessions = guardService.GetSessions("127.0.0.1")
		assert.Len(t, sessions, 2, "GetSessions should have length 2")

		res, _ = guardService.CheckAndRegisterSession("127.0.0.1", "Test", nil, "sandbox3")
		assert.True(t, res, "CheckAndRegisterSession should return true")
		sessions = guardService.GetSessions("127.0.0.1")
		assert.Len(t, sessions, 3, "GetSessions should have length 3")

		res, _ = guardService.CheckAndRegisterSession("127.0.0.1", "Test", nil, "sandbox4")
		assert.False(t, res, "CheckAndRegisterSession should return false")
		sessions = guardService.GetSessions("127.0.0.1")
		assert.Len(t, sessions, 3, "GetSessions should have length 3")

		guardService.UnregisterSession("sandbox2")

		res, _ = guardService.CheckAndRegisterSession("127.0.0.1", "Test", nil, "sandbox5")
		assert.True(t, res, "CheckAndRegisterSession should return true")
		sessions = guardService.GetSessions("127.0.0.1")
		assert.Len(t, sessions, 3, "GetSessions should have length 3")
	})

}
