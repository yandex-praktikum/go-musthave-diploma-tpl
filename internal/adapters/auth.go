package adapters

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sashaaro/go-musthave-diploma-tpl/internal/utils"
	"net/http"
)

type userContext struct{}

func UserIDFromReq(req *http.Request) (uuid.UUID, error) {
	return UserIDFromCtx(req.Context())
}
func MustUserIDFromReq(req *http.Request) uuid.UUID {
	return utils.Must(UserIDFromReq(req))
}

func UserIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(&userContext{}).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user id not found")
	}
	return userID, nil
}

func UserIDToCxt(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, &userContext{}, userID)
}
