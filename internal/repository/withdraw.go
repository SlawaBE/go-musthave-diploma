package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	BlockingUserForUpdate  = `SELECT 1 FROM users WHERE id = $1 FOR UPDATE`
	SelectWithdrawByUserID = `SELECT id, user_id, order_number, total, processed_at FROM withdrawns WHERE user_id = $1 ORDER BY processed_at DESC;`
	InsertWithdraw         = `
		WITH o AS (SELECT coalesce(sum(accrual), 0) as total FROM orders WHERE user_id = $1),
        	 w AS (SELECT coalesce(sum(total), 0) as total FROM withdrawns WHERE user_id = $1)
        INSERT INTO withdrawns (user_id, order_number, total)
		SELECT $1, $2, $3
		FROM o, w
        WHERE o.total - w.total >= $3;
    `
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
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
		return err
	}
	defer tx.Rollback()

	var dummy int
	err = tx.QueryRowContext(ctx,
		BlockingUserForUpdate,
		withdraw.UserID,
	).Scan(&dummy)
	if err != nil {
		logger.Log.Error("Error blocking user", zap.Error(err))
		return err
	}

	res, err := tx.ExecContext(ctx, InsertWithdraw, withdraw.UserID, withdraw.OrderNumber, withdraw.Total)
	if err != nil {
		logger.Log.Error("Error update balance", zap.Error(err))
		return err
	}
	affectedRows, err := res.RowsAffected()
	if err != nil {
		logger.Log.Error("Error update balance", zap.Error(err))
		return err
	}
	if affectedRows == 0 {
		err = fmt.Errorf("%w: user_id=%d", ErrInsufficientFunds, withdraw.UserID)
		logger.Log.Error("Balance not updated", zap.Error(err))
		return err
	}

	return tx.Commit()
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
