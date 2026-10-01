package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jackc/pgx/v5/pgconn"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) SaveUser(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) BuildJWTString(userID uint64) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func TestRegisterHandler_ServeHTTP(t *testing.T) {
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
			name:        "successful register",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.RegisterUserRequest{Login: "testuser", Password: "password123"},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("SaveUser", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
				token.On("BuildJWTString", mock.Anything).Return("test.jwt.token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedCookie: true,
		},
		{
			name:           "wrong request method",
			method:         http.MethodGet,
			contentType:    "application/json",
			body:           nil,
			setupMocks:     func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCookie: false,
		},
		{
			name:           "wrong Content-Type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			body:           nil,
			setupMocks:     func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusBadRequest,
			expectedCookie: false,
		},
		{
			name:           "empty login",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           model.RegisterUserRequest{Login: "", Password: "password123"},
			setupMocks:     func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusBadRequest,
			expectedCookie: false,
		},
		{
			name:           "empty password",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           model.RegisterUserRequest{Login: "testuser", Password: ""},
			setupMocks:     func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusBadRequest,
			expectedCookie: false,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			contentType:    "application/json",
			body:           "{invalid json",
			setupMocks:     func(repo *MockUserRepository, token *MockTokenService) {},
			expectedStatus: http.StatusBadRequest,
			expectedCookie: false,
		},
		{
			name:        "user already exists",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.RegisterUserRequest{Login: "existinguser", Password: "password123"},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				pgErr := &pgconn.PgError{Code: "23505"}
				repo.On("SaveUser", mock.Anything, mock.AnythingOfType("*model.User")).Return(pgErr)
			},
			expectedStatus: http.StatusConflict,
			expectedCookie: false,
		},
		{
			name:        "database error",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.RegisterUserRequest{Login: "testuser", Password: "password123"},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("SaveUser", mock.Anything, mock.AnythingOfType("*model.User")).Return(errors.New("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCookie: false,
		},
		{
			name:        "error creation JWT",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        model.RegisterUserRequest{Login: "testuser", Password: "password123"},
			setupMocks: func(repo *MockUserRepository, token *MockTokenService) {
				repo.On("SaveUser", mock.Anything, mock.AnythingOfType("*model.User")).Return(nil)
				token.On("BuildJWTString", mock.Anything).Return("", errors.New("jwt generation error"))
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

			handler := &RegisterHandler{
				repository:   repo,
				tokenService: token,
			}

			var bodyBytes []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					bodyBytes = []byte(str)
				} else {
					bodyBytes, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/register", bytes.NewReader(bodyBytes))
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

func TestIsNotUniqError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "PostgreSQL not unique error",
			err:      &pgconn.PgError{Code: "23505"},
			expected: true,
		},
		{
			name:     "PostgreSQL another error",
			err:      &pgconn.PgError{Code: "23502"},
			expected: false,
		},
		{
			name:     "common error",
			err:      errors.New("common error"),
			expected: false,
		},
		{
			name:     "nil",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotUniqError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
