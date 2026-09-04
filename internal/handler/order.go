package handler

import (
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
	accrualService *service.AccrualService
}

func NewOrdersUploadHandler(repository *repository.OrderRepository, accrualService *service.AccrualService) *OrdersUploadHandler {
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

	// asr, err := h.accrualService.GetAccrual(orderNumber)
	// if err != nil {
	// 	http.Error(w, "Error getting accrual", http.StatusInternalServerError)
	// 	return
	// }

	order := &model.Order{
		Number:  orderNumber,
		UserID:  userID,
		Status:  "NEW",
	}
	err = h.repository.SaveOrder(r.Context(), *order)
	if err != nil {
		if IsNotUniqError(err) {
			order, err = h.repository.GetOrderByNumber(r.Context(), orderNumber)
			if err != nil {
				http.Error(w, "Error check order", http.StatusInternalServerError)
				return
			}
			if order.UserID == userID {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//TODO надо создать воркеров, которые в фоне будет обрабатывать заказы
	// - передавать новые заказы через канал отсюда
	// - другая горутина раз в единицу времени выгребает пачку NEW/PROCESSING и также передаёт через канал
	// - воркеры ходят в accrual
	// Обработка
	// - для PROCESSED/INVALID вышли
	// - для NEW поставили PROCESSING
	// - сходили в accrual-service
	// - для PROCESSED/INVALID обновили заказ поставили статус вернувшийся оттуда и сумму (если была)
	// - 429, 500 - обработается resty - повторный поход в рамках текущей задачи, при исчерпании попыток останется PROCESSING
	// - для 204 - будем считать, что необходимо прити позже (заказ догрузят) - тоже оставляем PROCESSING
	// Дополнение
	// - по идее, надо гребсти (NEW, PROCESSING) т.к. могут быть незавершённые задачи из-за перезапусков сервиса
	// - обработать заказ дважды не страшно т.к. баланс считается суммированием по заказам
	// - можно проверять наличие заказов в статусе PROCESSING при старте приложения и сбрасывать их в NEW
	// - было бы прикольно для задач не в финальном статусе снова записывать их в канал, но если он забъётся, то воркеры могут встать пытаясь записать в канал, из которого читают сами же

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}
