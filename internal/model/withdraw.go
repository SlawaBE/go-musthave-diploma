package model

import "github.com/jackc/pgx/v5/pgtype"

type Withdraw struct {
	ID          uint64
	UserID      uint64
	OrderNumber string
	Total       float32
	ProcessedAt pgtype.Timestamptz
}

type WithdrawRequest struct {
	OrderNumber string  `json:"order"`
	Total       float32 `json:"sum"`
}

type WithdrawItem struct {
	OrderNumber string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}
