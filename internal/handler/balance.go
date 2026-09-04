package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
)

type BalanceHandler struct {
	repository          *repository.OrderRepository
	userRepository      *repository.UserRepository
	withdrawnRepository *repository.WitdrawnRepository
}

func NewBalanceHandler(repository *repository.OrderRepository, userRepository *repository.UserRepository, withdrawnRepository *repository.WitdrawnRepository) *BalanceHandler {
	return &BalanceHandler{
		repository:          repository,
		userRepository:      userRepository,
		withdrawnRepository: withdrawnRepository,
	}
}

func (h *BalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	login, ok := service.GetLoginFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepository.GetUserByLogin(r.Context(), login)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	total, err := h.repository.GetSumOfAccrual(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error calc balance", http.StatusInternalServerError)
		return
	}
	withdrawn, err := h.withdrawnRepository.GetSumOfWithdrawn(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error calc balance", http.StatusInternalServerError)
		return
	}

	balance := &model.BalanceResponse{
		Current:   *total - *withdrawn,
		Withdrawn: *withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, "Error Marshalling response", http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
