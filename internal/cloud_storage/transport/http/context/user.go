package context

import (
	"context"
	"net/http"

	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

var userCtxKey = &contextKey{"user"}

func NewUserContext(ctx context.Context, u *models.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

func UserFromContext(ctx context.Context) *models.User {
	return ctx.Value(userCtxKey).(*models.User)
}

func UserFromRequest(r *http.Request) *models.User {
	return UserFromContext(r.Context())
}
