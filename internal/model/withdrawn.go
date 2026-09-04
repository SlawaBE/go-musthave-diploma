package model

import "github.com/jackc/pgx/v5/pgtype"

type Withdrawn struct {
	ID          uint64
	UserID      uint64
	OrderNumber string
	Total       float32
	ProcessedAt pgtype.Timestamptz
}
