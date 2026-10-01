package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenService(t *testing.T) {
	secretKey := "test-secret"
	tokenLifetime := 24 * time.Hour

	service := NewTokenService(secretKey, tokenLifetime)

	assert.NotNil(t, service)
	assert.Equal(t, secretKey, service.secretKey)
	assert.Equal(t, tokenLifetime, service.tokenLifetime)
}

func TestTokenService_BuildJWTString(t *testing.T) {
	tests := []struct {
		name          string
		userID        uint64
		secretKey     string
		tokenLifetime time.Duration
		wantErr       bool
	}{
		{
			name:          "successful token creation",
			userID:        12345,
			secretKey:     "test-secret",
			tokenLifetime: 24 * time.Hour,
			wantErr:       false,
		},
		{
			name:          "zero user ID",
			userID:        0,
			secretKey:     "test-secret",
			tokenLifetime: 24 * time.Hour,
			wantErr:       false,
		},
		{
			name:          "max user ID",
			userID:        18446744073709551615,
			secretKey:     "test-secret",
			tokenLifetime: 24 * time.Hour,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewTokenService(tt.secretKey, tt.tokenLifetime)
			tokenString, err := service.BuildJWTString(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, tokenString)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tokenString)

				claims, err := service.ValidateJWTString(tokenString)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
			}
		})
	}
}

func TestTokenService_ValidateJWTString(t *testing.T) {
	secretKey := "test-secret"
	tokenLifetime := 24 * time.Hour
	service := NewTokenService(secretKey, tokenLifetime)
	userID := uint64(12345)

	t.Run("valid token", func(t *testing.T) {
		tokenString, err := service.BuildJWTString(userID)
		require.NoError(t, err)

		claims, err := service.ValidateJWTString(tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.NotNil(t, claims.ExpiresAt)
	})

	t.Run("expired token", func(t *testing.T) {
		shortLifetimeService := NewTokenService(secretKey, 1*time.Millisecond)
		tokenString, err := shortLifetimeService.BuildJWTString(userID)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		claims, err := shortLifetimeService.ValidateJWTString(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("invalid token string", func(t *testing.T) {
		claims, err := service.ValidateJWTString("invalid-token-string")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "error parsing jwt token")
	})

	t.Run("empty token string", func(t *testing.T) {
		claims, err := service.ValidateJWTString("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("token signed with different secret", func(t *testing.T) {
		differentService := NewTokenService("different-secret", tokenLifetime)
		tokenString, err := differentService.BuildJWTString(userID)
		require.NoError(t, err)

		claims, err := service.ValidateJWTString(tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

type differentContextKey struct{}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantID    uint64
		wantFound bool
	}{
		{
			name:      "user ID exists in context",
			ctx:       context.WithValue(context.Background(), userIDContextKey{}, uint64(12345)),
			wantID:    12345,
			wantFound: true,
		},
		{
			name:      "user ID exists with zero value",
			ctx:       context.WithValue(context.Background(), userIDContextKey{}, uint64(0)),
			wantID:    0,
			wantFound: true,
		},
		{
			name:      "user ID not in context",
			ctx:       context.Background(),
			wantID:    0,
			wantFound: false,
		},
		{
			name:      "different key in context",
			ctx:       context.WithValue(context.Background(), differentContextKey{}, uint64(12345)),
			wantID:    0,
			wantFound: false,
		},
		{
			name:      "wrong type in context",
			ctx:       context.WithValue(context.Background(), userIDContextKey{}, "not-a-uint64"),
			wantID:    0,
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, found := GetUserIDFromContext(tt.ctx)
			assert.Equal(t, tt.wantFound, found)
			assert.Equal(t, tt.wantID, gotID)
		})
	}
}
