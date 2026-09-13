package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/go-resty/resty/v2"
)

type AccrualService struct {
	client          *resty.Client
	workersCount    int
	jobsSize        int
	orderRepository *repository.OrderRepository
	jobs            chan uint64
	wg              sync.WaitGroup
	pollInterval    int
	pollLimit       int
}

func NewAccrualService(accrualSystemAddress string, orderRepository *repository.OrderRepository) *AccrualService {
	return &AccrualService{
		client:          httpClient(accrualSystemAddress),
		workersCount:    5, //TODO конфигурация воркеров, размера канала, интервала и лимита поллинга
		jobs:            make(chan uint64, 100),
		orderRepository: orderRepository,
		pollInterval:    60,
		pollLimit:       100,
	}
}

type AccrualServiceResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float32 `json:"accrual,omitempty"`
}

func (a *AccrualService) Run(ctx context.Context) {
	for range a.workersCount {
		a.wg.Go(func() {
			a.work(ctx, a.jobs)
			logger.Log.Info("Accrual service worker stopped")
		})
	}

	a.wg.Go(func() {
		a.runPollerOldOrders(ctx)
		logger.Log.Info("Accrual service poller stopped")
	})
	logger.Log.Info("Accrual service started")
}

func (a *AccrualService) AddOrder(ctx context.Context, orderID uint64) {
	select {
	case a.jobs <- orderID:
	case <-ctx.Done():
	}
}

func (a *AccrualService) Stop() {
	logger.Log.Info("Stop accrual service")
	close(a.jobs)
	a.wg.Wait()
	logger.Log.Info("Accrual service stopped")
}

func (a *AccrualService) runPollerOldOrders(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			ids, err := a.orderRepository.GetNOldestNotProcessedOrderIDs(ctx, a.pollLimit)
			if err != nil {
				logger.Log.Error("error polling orders", logger.Err(err))
				continue
			}
			for _, id := range ids {
				select {
				case a.jobs <- id:
				case <-ctx.Done():
				}
			}
		}
	}
}

func (a *AccrualService) work(ctx context.Context, jobs <-chan uint64) {
	for {
		select {
		case <-ctx.Done():
			return
		case orderID, ok := <-jobs:
			if !ok {
				return
			}
			logger.Log.Info("Start processing order", slog.Uint64("order_id", orderID))
			status := a.process(ctx, orderID)
			logger.Log.Info("End processing order", slog.Uint64("order_id", orderID), slog.Any("status", status))
		}
	}
}

func (a *AccrualService) process(ctx context.Context, orderID uint64) model.OrderStatus {
	order, err := a.orderRepository.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Log.Error("error getting order", logger.Err(err))
		return ""
	}

	if order.Status == model.OrderStatusInvalid || order.Status == model.OrderStatusProcessed {
		logger.Log.Info("processing not requiered", slog.Uint64("order_id", orderID), slog.Any("status", order.Status))
		return order.Status
	}

	err = a.orderRepository.UpdateStatus(ctx, orderID, model.OrderStatusProcessing)
	if err != nil {
		logger.Log.Error("error getting order", logger.Err(err))
		return ""
	}

	asr, err := a.getAccrual(ctx, order.Number)
	if err != nil {
		logger.Log.Error("error getting accrual", logger.Err(err))
		return ""
	}

	switch asr.Status {
	case "REGISTERED", "PROCESSING":
		order.Status = model.OrderStatusProcessing
	case "PROCESSED":
		order.Accrual = asr.Accrual
		order.Status = model.OrderStatusProcessed
		err = a.orderRepository.SetAccrual(ctx, orderID, order.Accrual)
	case "INVALID":
		order.Status = model.OrderStatusInvalid
		err = a.orderRepository.UpdateStatus(ctx, order.ID, order.Status)
	}
	if err != nil {
		logger.Log.Error("error updating status", slog.Uint64("order_id", order.ID), slog.Any("status", order.Status))
		return ""
	}

	return order.Status
}

func (a *AccrualService) getAccrual(ctx context.Context, number string) (*AccrualServiceResponse, error) {
	resp, err := a.client.R().
		SetContext(ctx).
		Get(number)
	if err != nil {
		return nil, fmt.Errorf("error getting accrual: %v", err)
	}
	var asr AccrualServiceResponse
	if resp.IsError() {
		message := fmt.Sprintf("error getting accrual, status code=%d", resp.StatusCode())
		return nil, errors.New(message)
	}
	if resp.StatusCode() == 204 {
		message := fmt.Sprintf("order %s not registered", number)
		return nil, errors.New(message)
	}
	err = json.Unmarshal(resp.Body(), &asr)
	return &asr, err
}

var retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func httpClient(baseURL string) *resty.Client {
	client := resty.New().
		SetBaseURL(baseURL + "/api/orders/").
		SetRetryCount(len(retryIntervals)).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			return retryIntervals[r.Request.Attempt-1], nil
		}).
		SetRetryMaxWaitTime(retryIntervals[len(retryIntervals)-1])
	return client
}
