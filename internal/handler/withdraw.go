package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
	"github.com/SlawaBE/go-musthave-diploma/internal/utils/validator"
	"go.uber.org/zap"
)

type WitdrawUploadHandler struct {
	repository      *repository.WithdrawRepository
	userRepository  *repository.UserRepository
	orderRepository *repository.OrderRepository
}

func NewWitdrawUploadHandler(repository *repository.WithdrawRepository, userRepository *repository.UserRepository, orderRepository *repository.OrderRepository) *WitdrawUploadHandler {
	return &WitdrawUploadHandler{
		repository:      repository,
		userRepository:  userRepository,
		orderRepository: orderRepository,
	}
}

func (h *WitdrawUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		logger.Log.Error("cannot decode request JSON body", zap.Error(err))
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

	balance, err := h.orderRepository.GetSumOfAccrual(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error calc balance", http.StatusInternalServerError)
		return
	}

	withdrawals, err := h.repository.GetSumOfWithdraw(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error calc balance", http.StatusInternalServerError)
		return
	}

	if *balance-*withdrawals < request.Total {
		http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		return
	}

	withdraw := &model.Withdraw{
		OrderNumber: request.OrderNumber,
		UserID:      user.ID,
		Total:       request.Total,
	}
	err = h.repository.SaveWitdrawn(r.Context(), *withdraw)
	if err != nil {
		var message string
		if IsNotUniqError(err) {
			message = "order number has already been used"
		} else {
			message = "error save withdraw"
		}
		http.Error(w, message, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
