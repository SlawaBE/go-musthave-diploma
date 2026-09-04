package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type AccrualService struct {
	client *resty.Client
}

func NewAccrualService(accrualSystemAddress string) *AccrualService {
	return &AccrualService{
		client: httpClient(accrualSystemAddress),
	}
}

type AccrualServiceResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float32 `json:"accrual,omitempty"`
}

func (a *AccrualService) GetAccrual(number string) (*AccrualServiceResponse, error) {
	resp, err := a.client.R().Get(number)
	if err != nil {
		logger.Log.Error("error getting accrual", zap.Error(err))
		return nil, fmt.Errorf("error getting accrual: %v", err)
	}
	var asr AccrualServiceResponse
	if resp.IsError() {
		message := fmt.Sprintf("error getting accrual, status code: %d", resp.StatusCode())
		logger.Log.Error(message)
		return nil, errors.New(message)
	}
	if resp.StatusCode() == 204 {
		message := fmt.Sprintf("order %s not registered", number)
		logger.Log.Error(message)
		return nil, errors.New(message)
	}
	json.Unmarshal(resp.Body(), &asr)
	return &asr, nil
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
