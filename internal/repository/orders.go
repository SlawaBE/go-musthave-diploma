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
	InsertOrder              = `INSERT INTO orders (user_id, number, status, accrual) VALUES ($1, $2, $3, $4);`
	SelectOrderByNumber      = `SELECT id, user_id, number, status, uploaded_at, accrual FROM orders WHERE number = $1;`
	SelectOrderByUserID      = `SELECT id, user_id, number, status, uploaded_at, accrual FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC;`
	SelectSumAccrualByUserID = `SELECT coalesce(sum(accrual), 0) as total FROM orders WHERE user_id = $1;`
)

func (w *OrderRepository) SaveOrder(ctx context.Context, order model.Order) error {
	tx, err := w.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, InsertOrder)
	if err != nil {
		logger.Log.Error("error prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, order.UserID, order.Number, order.Status, order.Accrual)
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

func (w *OrderRepository) GetOrderByNumber(ctx context.Context, number string) (*model.Order, error) {
	var order model.Order
	row := w.db.QueryRowContext(ctx, SelectOrderByNumber, number)
	var err error

	if err = row.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.UploadedAt, &order.Accrual); err != nil {
		logger.Log.Error("error get order", zap.String("number", number), zap.Error(err))
		return nil, err
	}
	return &order, nil
}

func (w *OrderRepository) Orders(ctx context.Context, userID uint64) ([]model.Order, error) {
	orders := make([]model.Order, 0)
	rows, err := w.db.QueryContext(ctx, SelectOrderByUserID, userID)
	if err != nil {
		logger.Log.Error("error query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order model.Order
		if err = rows.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.UploadedAt, &order.Accrual); err != nil {
			logger.Log.Error("error get order", zap.Uint64("userId", userID), zap.Error(err))
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

func (w *OrderRepository) GetSumOfAccrual(ctx context.Context, userID uint64) (*float32, error) {
	row := w.db.QueryRowContext(ctx, SelectSumAccrualByUserID, userID)

	var sum float32
	if err := row.Scan(&sum); err != nil {
		logger.Log.Error("error sum accrual", zap.Uint64("userId", userID), zap.Error(err))
		return nil, err
	}

	return &sum, nil
}
