package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type RegisterHandler struct {
	repository   *repository.UserRepository
	tokenService *service.TokenService
}

func NewRegisterHandler(repository *repository.UserRepository, tokenService *service.TokenService) *RegisterHandler {
	return &RegisterHandler{
		repository:   repository,
		tokenService: tokenService,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var user model.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		logger.Log.Error("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if user.Login == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := h.repository.CreateUser(r.Context(), user)

	if err != nil {
		if IsNotUniqError(err) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwtToken, err := h.tokenService.BuildJWTString(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", "Bearer "+jwtToken)
	w.WriteHeader(http.StatusOK)
}

func IsNotUniqError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
