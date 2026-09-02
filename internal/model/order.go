package model

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Id         uint64
	UserId     uint64
	Number     string
	Status     OrderStatus
	UploadedAt pgtype.Timestamptz
}

func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}

	switch v := value.(type) {
	case string:
		*s = OrderStatus(v)
		return nil
	case []byte:
		*s = OrderStatus(string(v))
		return nil
	default:
		return fmt.Errorf("cannot scan %T into OrderStatus", value)
	}
}

func (s OrderStatus) Value() (driver.Value, error) {
	return string(s), nil
}
