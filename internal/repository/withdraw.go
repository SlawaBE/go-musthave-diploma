package repository

import (
	"context"
	"database/sql"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

type WitdrawRepository struct {
	db *sql.DB
}

func NewWitdrawRepository(db *sql.DB) *WitdrawRepository {
	return &WitdrawRepository{
		db: db,
	}
}

const (
	SelectSumTotalByUserID = `SELECT coalesce(sum(total), 0) as total FROM withdrawns WHERE user_id = $1;`
	InsertWithdraw         = `INSERT INTO withdrawns (user_id, order_number, total) VALUES ($1, $2, $3);`
)

func (w *WitdrawRepository) GetSumOfWithdraw(ctx context.Context, userID uint64) (*float32, error) {
	row := w.db.QueryRowContext(ctx, SelectSumTotalByUserID, userID)

	var sum float32
	if err := row.Scan(&sum); err != nil {
		logger.Log.Error("error sum accrual", zap.Uint64("userId", userID), zap.Error(err))
		return nil, err
	}

	return &sum, nil
}

func (w *WitdrawRepository) SaveWitdrawn(ctx context.Context, withdraw model.Withdraw) error {
	tx, err := w.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, InsertWithdraw)
	if err != nil {
		logger.Log.Error("error prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, withdraw.UserID, withdraw.OrderNumber, withdraw.Total)
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
