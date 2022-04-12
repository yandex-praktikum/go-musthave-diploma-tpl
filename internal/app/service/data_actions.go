package service

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"github.com/EestiChameleon/BonusServiceGOphermart/internal/app/cfg"
	"github.com/robbert229/jwt"
)

var (
	ErrInvalidToken = errors.New("failed to decode the provided Token")
)

func EncryptPass(pass string) (string, error) {
	block, err := aes.NewCipher([]byte(cfg.Envs.CryptoKey))
	if err != nil {
		return "", err
	}
	plainText := []byte(pass)
	cfb := cipher.NewCFBEncrypter(block, []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05})
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func JWTEncode(key string, value interface{}) (string, error) {
	algorithm := jwt.HmacSha256(cfg.Envs.CryptoKey)

	claims := jwt.NewClaim()
	claims.Set(key, value)

	token, err := algorithm.Encode(claims)
	if err != nil {
		return ``, err
	}

	if err = algorithm.Validate(token); err != nil {
		return ``, err
	}

	return token, nil
}

func JWTDecode(token, key string) (interface{}, error) {
	algorithm := jwt.HmacSha256(cfg.Envs.CryptoKey)

	if err := algorithm.Validate(token); err != nil {
		return nil, err
	}

	claims, err := algorithm.Decode(token)
	if err != nil {
		return nil, err
	}

	return claims.Get(key)
}
