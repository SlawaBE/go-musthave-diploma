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
	InsertOrder                    = `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3) RETURNING id;`
	SelectOrderByNumber            = `SELECT id, user_id, number, status, uploaded_at, accrual FROM orders WHERE number = $1;`
	SelectOrderByUserID            = `SELECT id, user_id, number, status, uploaded_at, accrual FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC;`
	SelectSumAccrualByUserID       = `SELECT coalesce(sum(accrual), 0) as total FROM orders WHERE user_id = $1;`
	SelectOrderByID                = `SELECT id, user_id, number, status, uploaded_at, accrual FROM orders WHERE id = $1;`
	UpdateOrderStatus              = `UPDATE orders SET status = $2 WHERE id = $1;`
	SetAccrual                     = `UPDATE orders SET accrual = $2, status = $3 WHERE id = $1;`
	GetNOldestNotProcessedOrderIDs = `SELECT id FROM orders WHERE status = 'NEW' or status = 'PROCESSING' ORDER BY uploaded_at ASC LIMIT $1`
)

func (w *OrderRepository) SaveOrder(ctx context.Context, order *model.Order) error {
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

	err = stmt.QueryRowContext(ctx, order.UserID, order.Number, order.Status).Scan(&order.ID)
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

func (w *OrderRepository) GetOrderByID(ctx context.Context, orderID uint64) (*model.Order, error) {
	var order model.Order
	row := w.db.QueryRowContext(ctx, SelectOrderByID, orderID)
	var err error

	if err = row.Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.UploadedAt, &order.Accrual); err != nil {
		logger.Log.Error("error get order", zap.Uint64("order_id", orderID), zap.Error(err))
		return nil, err
	}
	return &order, nil
}

func (w *OrderRepository) UpdateStatus(ctx context.Context, orderID uint64, status model.OrderStatus) error {
	tx, err := w.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, UpdateOrderStatus)
	if err != nil {
		logger.Log.Error("error prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, orderID, status)
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

func (w *OrderRepository) SetAccrual(ctx context.Context, orderID uint64, accrual *float32) error {
	tx, err := w.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, SetAccrual)
	if err != nil {
		logger.Log.Error("error prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, orderID, accrual, model.OrderStatusProcessed)
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

func (w *OrderRepository) GetNOldestNotProcessedOrderIDs(ctx context.Context, limit int) ([]uint64, error) {
	ids := make([]uint64, 0)
	rows, err := w.db.QueryContext(ctx, GetNOldestNotProcessedOrderIDs, limit)
	if err != nil {
		logger.Log.Error("error query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		if err = rows.Scan(&id); err != nil {
			logger.Log.Error("error get order", zap.Error(err))
			return nil, err
		}
		ids = append(ids, id)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("error get orders", zap.Error(err))
		return nil, err
	}
	return ids, nil
}
