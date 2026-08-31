package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}
