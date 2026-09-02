package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"go.uber.org/zap"
)

func Sha256(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hasher.Sum(nil)
}

func Check(data []byte, hash []byte) bool {
	hashOfData := Sha256(data)
	return hmac.Equal(hash, hashOfData)
}

func CheckPassword(password string, hash string) bool {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		logger.Log.Error("wrong password hash", zap.Error(err))
		return false
	}
	return Check([]byte(password), decodedHash)
}
