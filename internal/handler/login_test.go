package handler

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/utils/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestLoginHandler_ServeHTTP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	testuser := "testuser"
	password := "password123"
	passwordHash := hex.EncodeToString(hash.Sha256([]byte(password)))

	tests := []struct {
		name           string
		method         string
		contentType    string
		body           interface{}
		setupMocks     func(repo *MockUserRepository, token *MockTokenService)
		expectedStatus int
		expectedCookie bool
	}{
		{
			name:        "successful login",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.LoginUserRequest{Login: testuser, Password: password},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("GetUserByLogin", mock.Anything, testuser).
					Return(&model.User{
						ID:           123,
						Login:        testuser,
						PasswordHash: passwordHash,
					}, nil)
				token.On("BuildJWTString", uint64(123)).Return("test.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedCookie: true,
		},
		{
			name:        "wrong request method",
			method:      http.MethodGet,
			contentType: "application/json",
			body:        nil,
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCookie: false,
		},
		{
			name:        "wrong Content-Type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        nil,
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusBadRequest,
			expectedCookie: false,
		},
		{
			name:        "empty login",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.LoginUserRequest{Login: "", Password: password},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusBadRequest,
			expectedCookie: false,
		},
		{
			name:        "invalid JSON",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{invalid json",
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusBadRequest,
			expectedCookie: false,
		},
		{
			name:        "user not found",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.LoginUserRequest{Login: "nonexistent", Password: password},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("GetUserByLogin", mock.Anything, "nonexistent").
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCookie: false,
		},
		{
			name:        "database error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.LoginUserRequest{Login: testuser, Password: password},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("GetUserByLogin", mock.Anything, testuser).
					Return(nil, errors.New("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCookie: false,
		},
		{
			name:        "wrong пароль",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.LoginUserRequest{Login: testuser, Password: "wrongpassword"},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("GetUserByLogin", mock.Anything, testuser).
					Return(&model.User{
						ID:           123,
						Login:        testuser,
						PasswordHash: passwordHash,
					}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedCookie: false,
		},
		{
			name:        "error creation JWT",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.LoginUserRequest{Login: testuser, Password: password},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("GetUserByLogin", mock.Anything, testuser).
					Return(&model.User{
						ID:           123,
						Login:        testuser,
						PasswordHash: passwordHash,
					}, nil)
				token.On("BuildJWTString", uint64(123)).Return("", errors.New("jwt generation error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCookie: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			token := new(MockTokenService)
			tt.setupMocks(repo, token)

			handler := NewLoginHandler(repo, token)

			var bodyBytes []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					bodyBytes = []byte(str)
				} else {
					bodyBytes, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			cookies := w.Result().Cookies()
			if tt.expectedCookie {
				assert.Len(t, cookies, 1)
				assert.Equal(t, "token", cookies[0].Name)
				assert.Equal(t, "test.jwt.token", cookies[0].Value)
				assert.True(t, cookies[0].HttpOnly)
				assert.Equal(t, "/", cookies[0].Path)
				assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
				assert.Equal(t, 1800, cookies[0].MaxAge)
			} else {
				assert.Len(t, cookies, 0)
			}

			repo.AssertExpectations(t)
			token.AssertExpectations(t)
		})
	}
}
