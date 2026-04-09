package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"github.com/mr-pixel-kg/shopshredder/api/internal/types"
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

func AuthFromContext(r *http.Request) *types.AuthContext {
	auth, ok := r.Context().Value(authKey).(types.AuthContext)
	if !ok {
		return nil
	}
	return &auth
}

func UserFromContext(r *http.Request) *models.User {
	u, _ := r.Context().Value(userKey).(*models.User)
	return u
}

func ClientIDFromContext(r *http.Request) *uuid.UUID {
	id, _ := r.Context().Value(clientIDKey).(*uuid.UUID)
	return id
}
