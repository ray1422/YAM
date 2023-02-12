package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func hmacSha256Str(data string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
