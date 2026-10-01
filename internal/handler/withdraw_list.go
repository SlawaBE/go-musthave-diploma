package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
)

type WithdrawListHandler struct {
	repository     WithdrawRepository
}

func NewWithdrawListHandler(repository WithdrawRepository) *WithdrawListHandler {
	return &WithdrawListHandler{
		repository:     repository,
	}
}

func (h *WithdrawListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := service.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dbWithdrawals, err := h.repository.Withdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error check order", http.StatusInternalServerError)
		return
	}

	if len(dbWithdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	orders := make([]model.WithdrawItem, len(dbWithdrawals))

	for i, v := range dbWithdrawals {
		orders[i] = model.WithdrawItem{
			OrderNumber: v.OrderNumber,
			Sum:         v.Total,
			ProcessedAt: v.ProcessedAt.Time.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json, err := json.Marshal(orders)
	if err != nil {
		http.Error(w, "Error Marshalling response", http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
