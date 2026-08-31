package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenService struct {
	secretKey     string
	tokenLifetime time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

func NewTokenService(secretKey string, tokenLifetime time.Duration) *TokenService {
	return &TokenService{
		secretKey:     secretKey,
		tokenLifetime: tokenLifetime,
	}
}

func (s *TokenService) BuildJWTString(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenLifetime)),
		},
		Login: login,
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("error signed jwt string %v", err)
	}

	return tokenString, nil
}
