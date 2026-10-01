package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
	"github.com/SlawaBE/go-musthave-diploma/internal/utils/validator"
)

type WithdrawUploadHandler struct {
	repository      WithdrawRepository
	orderRepository OrderRepository
}

func NewWithdrawUploadHandler(repository WithdrawRepository, orderRepository OrderRepository) *WithdrawUploadHandler {
	return &WithdrawUploadHandler{
		repository:      repository,
		orderRepository: orderRepository,
	}
}

func (h *WithdrawUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var request model.WithdrawRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		logger.Log.Error("cannot decode request JSON body", logger.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderNumber := request.OrderNumber

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

	withdraw := &model.Withdraw{
		OrderNumber: request.OrderNumber,
		UserID:      userID,
		Total:       request.Total,
	}
	err := h.repository.SaveWithdraw(r.Context(), *withdraw)
	switch {
	case errors.Is(err, repository.ErrInsufficientFunds):
		http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		return
	case IsNotUniqError(err):
		http.Error(w, "withdraw exists yet", http.StatusConflict)
		return
	case err != nil:
		logger.Log.Error("Error update balance", logger.Err(err))
		http.Error(w, "Error update balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
