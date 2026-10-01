package service

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateCode() (string, error) {
	data := make([]byte, 9)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}
