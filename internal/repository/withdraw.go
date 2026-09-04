package repository

import (
	"context"
	"database/sql"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

type WithdrawRepository struct {
	db *sql.DB
}

func NewWitdrawRepository(db *sql.DB) *WithdrawRepository {
	return &WithdrawRepository{
		db: db,
	}
}

const (
	SelectSumTotalByUserID = `SELECT coalesce(sum(total), 0) as total FROM withdrawns WHERE user_id = $1;`
	InsertWithdraw         = `INSERT INTO withdrawns (user_id, order_number, total) VALUES ($1, $2, $3);`
	SelectWithdrawByUserID = `SELECT id, user_id, order_number, total, processed_at FROM withdrawns WHERE user_id = $1 ORDER BY processed_at DESC;`
)

func (w *WithdrawRepository) GetSumOfWithdraw(ctx context.Context, userID uint64) (*float32, error) {
	row := w.db.QueryRowContext(ctx, SelectSumTotalByUserID, userID)

	var sum float32
	if err := row.Scan(&sum); err != nil {
		logger.Log.Error("error sum accrual", zap.Uint64("userId", userID), zap.Error(err))
		return nil, err
	}

	return &sum, nil
}

func (w *WithdrawRepository) SaveWitdrawn(ctx context.Context, withdraw model.Withdraw) error {
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

func (w *WithdrawRepository) Withdrawals(ctx context.Context, userID uint64) ([]model.Withdraw, error) {
	withdrawals := make([]model.Withdraw, 0)
	rows, err := w.db.QueryContext(ctx, SelectWithdrawByUserID, userID)
	if err != nil {
		logger.Log.Error("error query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdraw model.Withdraw
		if err = rows.Scan(&withdraw.ID, &withdraw.UserID, &withdraw.OrderNumber, &withdraw.Total, &withdraw.ProcessedAt); err != nil {
			logger.Log.Error("error get withdraw", zap.Uint64("userId", userID), zap.Error(err))
			return nil, err
		}
		withdrawals = append(withdrawals, withdraw)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("error get withdrawals", zap.Error(err))
		return nil, err
	}
	return withdrawals, nil
}
