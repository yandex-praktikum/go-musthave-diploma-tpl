package utils

import (
	"context"
)

// Сохранение id пользователя в контексте и получение из контекста
type userIDCtx struct{}

func SetUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDCtx{}, id)
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDCtx{}).(string); ok {
		return id
	}
	return ""
}
