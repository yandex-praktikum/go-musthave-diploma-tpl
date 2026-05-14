// Хэширование и проверка пароля

package utils

import (
	"github.com/alexedwards/argon2id"
)

type Hasher struct {
	params *argon2id.Params
}

func NewHasher(saltLen uint32, keyLen uint32) *Hasher {
	var defaultParams = &argon2id.Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  saltLen,
		KeyLength:   keyLen,
	}
	return &Hasher{
		params: defaultParams,
	}
}

func (h *Hasher) HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, h.params)
}

func (h *Hasher) VerifyPassword(password, encodedPassword string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, encodedPassword)
}
