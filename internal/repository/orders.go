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
	INSERT_ORDER = `INSERT INTO orders (user_id, number, status, accrual) VALUES ($1, $2, $3, $4);`
	SELECT_ORDER_BY_NUMBER = `SELECT id, user_id, number, status, uploaded_at, accrual FROM orders WHERE number = $1;`
	SELECT_ORDER_BY_USER_ID = `SELECT id, user_id, number, status, uploaded_at, accrual FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC;`
)

func (o *OrderRepository) SaveOrder(ctx context.Context, order model.Order) error {
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

	_, err = stmt.ExecContext(ctx, order.UserId, order.Number, order.Status, order.Accrual)
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

	if err = rows.Scan(&order.Id, &order.UserId, &order.Number, &order.Status, &order.UploadedAt, &order.Accrual); err != nil {
		logger.Log.Error("error get order", zap.String("number", number), zap.Error(err))
		return nil, err
	}
	return &order, nil
}

func (o *OrderRepository) Orders(ctx context.Context, userId uint64) ([]model.Order, error) {
	orders := make([]model.Order, 0)
	rows, err := o.db.QueryContext(ctx, SELECT_ORDER_BY_USER_ID, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var order model.Order
		if err = rows.Scan(&order.Id, &order.UserId, &order.Number, &order.Status, &order.UploadedAt, &order.Accrual); err != nil {
			logger.Log.Error("error get order", zap.Uint64("userId", userId), zap.Error(err))
			return nil, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("error get orders", zap.Error(err))
		return nil, err
	}
	return orders, nil
}
