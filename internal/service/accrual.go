package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
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
		return nil, fmt.Errorf("error getting accrual: %v", err)
	}
	var asr AccrualServiceResponse
	if resp.IsError() {
		return nil, fmt.Errorf("error getting accrual: %v", err)
	}
	if resp.StatusCode() == 204 {
		return nil, fmt.Errorf("order not registered")
	}
	json.Unmarshal(resp.Body(), &asr)
	return &asr, nil
}

var retryIntervals = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func httpClient(baseUrl string) *resty.Client {
	client := resty.New().
		SetBaseURL(baseUrl + "/api/orders/").
		SetRetryCount(len(retryIntervals)).
		SetRetryAfter(func(c *resty.Client, r *resty.Response) (time.Duration, error) {
			return retryIntervals[r.Request.Attempt-1], nil
		}).
		SetRetryMaxWaitTime(retryIntervals[len(retryIntervals)-1])
	return client
}
