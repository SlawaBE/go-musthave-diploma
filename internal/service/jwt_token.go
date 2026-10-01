package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenService struct {
	secretKey     string
	tokenLifetime time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uint64 `json:"user_id"`
}

func NewTokenService(secretKey string, tokenLifetime time.Duration) *TokenService {
	return &TokenService{
		secretKey:     secretKey,
		tokenLifetime: tokenLifetime,
	}
}

type userIDContextKey struct{}

func (s *TokenService) BuildJWTString(userID uint64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenLifetime)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", fmt.Errorf("error signed jwt string %v", err)
	}

	return tokenString, nil
}

func (s *TokenService) ValidateJWTString(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing jwt token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

func (s *TokenService) CreateAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
				return
			}

			claims, err := s.ValidateJWTString(cookie.Value)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unauthorized: %v", err), http.StatusUnauthorized)
				return
			}
			
			ctx := context.WithValue(r.Context(), userIDContextKey{}, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (uint64, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(uint64)
	return userID, ok
}
