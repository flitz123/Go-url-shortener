package service

import (
	"crypto/sha1"
	"encoding/base64"
)

func GenerateCode(url string) string {
	hash := sha1.Sum([]byte(url))
	return base64.URLEncoding.EncodeToString(hash[:])[:6]
}
