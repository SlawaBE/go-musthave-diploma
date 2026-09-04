package handler

import (
	"encoding/json"
	"net/http"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
)

type BalanceHandler struct {
	repository         *repository.OrderRepository
	withdrawRepository *repository.WithdrawRepository
}

func NewBalanceHandler(repository *repository.OrderRepository, withdrawRepository *repository.WithdrawRepository) *BalanceHandler {
	return &BalanceHandler{
		repository:         repository,
		withdrawRepository: withdrawRepository,
	}
}

func (h *BalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := service.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	total, err := h.repository.GetSumOfAccrual(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error calc balance", http.StatusInternalServerError)
		return
	}
	withdraw, err := h.withdrawRepository.GetSumOfWithdraw(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error calc balance", http.StatusInternalServerError)
		return
	}

	balance := &model.BalanceResponse{
		Current:  *total - *withdraw,
		Withdraw: *withdraw,
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
