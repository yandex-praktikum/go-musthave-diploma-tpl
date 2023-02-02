package events

import (
	"crypto/aes"
	"log"
	"time"
)

var key = append([]byte(time.Now().Format("02-01-2006 15")), []byte(":00")...)

func CryptoToken(token []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	cryptoString := make([]byte, len(token))
	aesBlock.Encrypt(cryptoString, token)
	return cryptoString, nil
}

func DeCryptoToken(token []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	deCryptoString := make([]byte, len(token))
	aesBlock.Decrypt(deCryptoString, token)
	return deCryptoString, nil
}
