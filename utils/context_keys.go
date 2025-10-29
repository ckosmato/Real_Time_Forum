package utils

import (
	"context"
	"real-time-forum/models"
)

type contextKey string

const ContextUser contextKey = "user"

func GetUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(ContextUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}