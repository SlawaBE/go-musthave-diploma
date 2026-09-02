package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
	"github.com/SlawaBE/go-musthave-diploma/internal/utils/hash"
	"go.uber.org/zap"
)

type LoginHandler struct {
	repository   *repository.UserRepository
	tokenService *service.TokenService
}

func NewLoginHandler(repository *repository.UserRepository, tokenService *service.TokenService) *LoginHandler {
	return &LoginHandler{
		repository:   repository,
		tokenService: tokenService,
	}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbUser, err := h.repository.GetUser(r.Context(), user.Login)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if !hash.CheckPassword(user.Password, dbUser.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	jwtToken, err := h.tokenService.BuildJWTString(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   false, //TODO конфигурировать
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   1800, //TODO конфигурировать таким же значением как TokenService.tokenLifetime (30 минут)
	})
	w.WriteHeader(http.StatusOK)
}
