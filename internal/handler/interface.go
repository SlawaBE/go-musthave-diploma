package handler

import (
	"context"

	"github.com/SlawaBE/go-musthave-diploma/internal/model"
)

type TokenService interface {
	BuildJWTString(userID uint64) (string, error)
}

type UserRepository interface {
	SaveUser(ctx context.Context, user *model.User) error
	GetUserByLogin(ctx context.Context, login string) (*model.User, error)
}

type WithdrawRepository interface {
	GetSumOfWithdraw(ctx context.Context, userID uint64) (*float32, error)
	SaveWitdrawn(ctx context.Context, withdraw model.Withdraw) error
	Withdrawals(ctx context.Context, userID uint64) ([]model.Withdraw, error)
}

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *model.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*model.Order, error)
	Orders(ctx context.Context, userID uint64) ([]model.Order, error)
	GetSumOfAccrual(ctx context.Context, userID uint64) (*float32, error)
	GetOrderByID(ctx context.Context, orderID uint64) (*model.Order, error)
	UpdateStatus(ctx context.Context, orderID uint64, status model.OrderStatus) error
	SetAccrual(ctx context.Context, orderID uint64, accrual *float32) error
	GetNOldestNotProcessedOrderIDs(ctx context.Context, limit int) ([]uint64, error)
}
