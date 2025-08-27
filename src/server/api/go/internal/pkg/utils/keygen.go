package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateKey
// A prefix can be passed in to generate a random string.
func GenerateKey(prefix string) (string, error) {
	var sb strings.Builder
	sb.WriteString(prefix)

	for range 48 {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(base62Chars[num.Int64()])
	}

	return sb.String(), nil
}
