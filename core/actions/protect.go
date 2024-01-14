package actions

import (
	"crypto/sha256"
	"encoding/base64"
)

// @TODO move to env or db
const salt = "some_SAlt(*1_"

func protect(p string) string {
	h := sha256.New()
	h.Write([]byte(p + salt))
	hashedPassword := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashedPassword)
}
