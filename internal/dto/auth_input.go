package dto

import "github.com/danielgtaylor/huma/v2"

// --- DTO ---
type AuthInput struct {
	Authorization string `header:"Authorization" doc:"Bearer <token>"`
}

// Реализуем Resolver, чтобы проверять токен прямо при парсинге входа
func (i *AuthInput) Resolve(ctx huma.Context, prefix string) []error {
	if i.Authorization == "" {
		return []error{huma.Error401Unauthorized("Missing Authorization header")}
	}

	return nil
}
