package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/types"
)

type contextKey string

const (
	authKey     contextKey = "auth"
	userKey     contextKey = "user"
	clientIDKey contextKey = "client_id"
)

func withAuth(ctx context.Context, auth types.AuthContext) context.Context {
	return context.WithValue(ctx, authKey, auth)
}

func withUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func withClientID(ctx context.Context, id *uuid.UUID) context.Context {
	return context.WithValue(ctx, clientIDKey, id)
}

func MustAuth(r *http.Request) types.AuthContext {
	return r.Context().Value(authKey).(types.AuthContext)
}

func UserFromContext(r *http.Request) *models.User {
	u, _ := r.Context().Value(userKey).(*models.User)
	return u
}

func ClientIDFromContext(r *http.Request) *uuid.UUID {
	id, _ := r.Context().Value(clientIDKey).(*uuid.UUID)
	return id
}
