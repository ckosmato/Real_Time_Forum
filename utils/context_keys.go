package utils

import (
	"context"
	"real-time-forum/models"
)

type contextKey string

const (
	ContextUser contextKey = "user"
	ContextRole contextKey = "role"
)

func GetUserFromContext(ctx context.Context) *models.User {
	userRaw := ctx.Value(ContextUser)
	user, ok := userRaw.(*models.User)
	if !ok || user == nil {
		return nil
	}
	return user
}
