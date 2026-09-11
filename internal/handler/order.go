package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
	"github.com/SlawaBE/go-musthave-diploma/internal/utils/validator"
)

type OrdersUploadHandler struct {
	repository     OrderRepository
	accrualService *service.AccrualService
}

func NewOrdersUploadHandler(repository OrderRepository, accrualService *service.AccrualService) *OrdersUploadHandler {
	return &OrdersUploadHandler{
		repository:     repository,
		accrualService: accrualService,
	}
}

func (h *OrdersUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.ParseInt(orderNumber, 10, 64); err != nil {
		http.Error(w, "Invalid order number format", http.StatusBadRequest)
		return
	}

	if !validator.ValidateLuhn(orderNumber) {
		http.Error(w, "Invalid order number (Luhn check failed)", http.StatusUnprocessableEntity)
		return
	}

	userID, ok := service.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	order := &model.Order{
		Number: orderNumber,
		UserID: userID,
		Status: "NEW",
	}
	err = h.repository.SaveOrder(r.Context(), order)
	if err != nil {
		if IsNotUniqError(err) {
			order, err = h.repository.GetOrderByNumber(r.Context(), orderNumber)
			if err != nil {
				http.Error(w, "Error check order", http.StatusInternalServerError)
				return
			}
			if order.UserID == userID {
				if (order.Status == model.OrderStatusNew || order.Status == model.OrderStatusProcessing) {
					h.accrualService.AddOrder(order.ID)
				}
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.accrualService.AddOrder(order.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}
