package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenServiceGenerateAndParseUserToken(t *testing.T) {
	t.Parallel()

	service := NewTokenService(config.AuthConfig{
		JWTSecret:          "unit-test-secret",
		JWTTTLMinutes:      30,
		GuestJWTTTLMinutes: 120,
	})
	userID := uuid.New()

	token, tokenID, expiresAt, err := service.Generate(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, tokenID)
	assert.WithinDuration(t, time.Now().UTC().Add(30*time.Minute), expiresAt, 5*time.Second)

	claims, err := service.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims.UserID)
	assert.Equal(t, "user", claims.SessionType)
	assert.Equal(t, tokenID, claims.TokenID)
	assert.Equal(t, userID.String(), claims.Subject)
}

func TestTokenServiceGenerateAndParseGuestToken(t *testing.T) {
	t.Parallel()

	service := NewTokenService(config.AuthConfig{
		JWTSecret:          "unit-test-secret",
		JWTTTLMinutes:      30,
		GuestJWTTTLMinutes: 120,
	})
	sessionID := uuid.New()

	token, tokenID, expiresAt, err := service.GenerateGuest(sessionID)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, tokenID)
	assert.WithinDuration(t, time.Now().UTC().Add(120*time.Minute), expiresAt, 5*time.Second)

	claims, err := service.Parse(token)
	require.NoError(t, err)
	assert.Empty(t, claims.UserID)
	assert.Equal(t, "guest", claims.SessionType)
	assert.Equal(t, tokenID, claims.TokenID)
	assert.Equal(t, sessionID.String(), claims.Subject)
}

func TestTokenServiceParseRejectsTamperedToken(t *testing.T) {
	t.Parallel()

	service := NewTokenService(config.AuthConfig{
		JWTSecret:          "unit-test-secret",
		JWTTTLMinutes:      30,
		GuestJWTTTLMinutes: 120,
	})

	token, _, _, err := service.Generate(uuid.New())
	require.NoError(t, err)

	tampered := token + "tampered"
	claims, err := service.Parse(tampered)
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenServiceParseRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	service := NewTokenService(config.AuthConfig{JWTSecret: "unit-test-secret"})

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:      uuid.NewString(),
		SessionType: "user",
		TokenID:     uuid.NewString(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC().Add(-2 * time.Minute)),
		},
	})

	token, err := expired.SignedString([]byte("unit-test-secret"))
	require.NoError(t, err)

	claims, err := service.Parse(token)
	require.Error(t, err)
	assert.Nil(t, claims)
}
