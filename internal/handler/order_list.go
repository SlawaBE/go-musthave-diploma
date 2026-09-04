package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
)

type OrdersListHandler struct {
	repository     *repository.OrderRepository
	userRepository *repository.UserRepository
}

func NewOrdersListHandler(repository *repository.OrderRepository, userRepository *repository.UserRepository) *OrdersListHandler {
	return &OrdersListHandler{
		repository:     repository,
		userRepository: userRepository,
	}
}

func (h *OrdersListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	dbOrders, err := h.repository.Orders(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error check order", http.StatusInternalServerError)
		return
	}

	orders := make([]model.OrderItem, len(dbOrders))

	for i, v := range dbOrders {
		orders[i] = model.OrderItem{
			Number:     v.Number,
			Status:     v.Status,
			Accrual:    v.Accrual,
			UploadedAt: v.UploadedAt.Time.Format(time.RFC3339),
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
