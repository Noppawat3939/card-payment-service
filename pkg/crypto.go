package pkg

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashSHA256(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}
