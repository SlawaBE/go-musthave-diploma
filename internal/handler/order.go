package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
	"github.com/SlawaBE/go-musthave-diploma/internal/utils/validator"
)

type OrdersUploadHandler struct {
	repository     *repository.OrderRepository
	userRepository *repository.UserRepository
}

func NewOrdersUploadHandler(repository *repository.OrderRepository, userRepository *repository.UserRepository) *OrdersUploadHandler {
	return &OrdersUploadHandler{
		repository:     repository,
		userRepository: userRepository,
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

	fmt.Println(body)

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	if _, err := strconv.ParseInt(orderNumber, 10, 64); err != nil {
		http.Error(w, "Invalid order number format", http.StatusBadRequest)
		return
	}

	// Проверяем по алгоритму Луна
	if !validator.ValidateLuhn(orderNumber) {
		http.Error(w, "Invalid order number (Luhn check failed)", http.StatusUnprocessableEntity)
		return
	}

	// Получаем логин из контекста (добавленный middleware)
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

	order := &model.Order{
		Number: orderNumber,
		UserId: user.Id,
		Status: "NEW",
	}
	err = h.repository.SaveOrder(r.Context(), *order)
	if err != nil {
		if IsNotUniqError(err) {
			order, err = h.repository.GetOrderByNumber(r.Context(), orderNumber)
			if err != nil {
				http.Error(w, "Error check order", http.StatusInternalServerError)
				return
			}
			if order.UserId == user.Id {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}
