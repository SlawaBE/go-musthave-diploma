package repository

import (
	"context"
	"database/sql"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

const (
	INSERT_ORDER = `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3);`
	SELECT_ORDER_BY_NUMBER = `SELECT id, user_id, number, status, uploaded_at FROM orders WHERE number = $1`
)

func (o *OrderRepository) SaveOrder(ctx context.Context, Order model.Order) error {
	tx, err := o.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, INSERT_ORDER)
	if err != nil {
		logger.Log.Error("error prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, Order.UserId, Order.Number, Order.Status)
	if err != nil {
		logger.Log.Error("error exec statement", zap.Error(err))
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Error("error commit transaction", zap.Error(err))
	}
	return err
}

func (o *OrderRepository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	var order model.Order
	rows := o.db.QueryRowContext(ctx, SELECT_ORDER_BY_NUMBER, number)
	var err error

	if err = rows.Scan(&order.Id, &order.UserId, &order.Number, &order.Status, &order.UploadedAt); err != nil {
		logger.Log.Error("error get login", zap.String("number", number), zap.Error(err))
		return nil, err
	}
	return &order, nil
}
