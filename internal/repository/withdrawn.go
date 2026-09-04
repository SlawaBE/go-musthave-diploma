package repository

import (
	"context"
	"database/sql"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"go.uber.org/zap"
)

type WitdrawnRepository struct {
	db *sql.DB
}

func NewWitdrawnRepository(db *sql.DB) *WitdrawnRepository {
	return &WitdrawnRepository{
		db: db,
	}
}

const (
	SelectSumTotalByUserID = `SELECT coalesce(sum(total), 0) as total FROM withdrawns WHERE user_id = $1;`
)

func (o *WitdrawnRepository) GetSumOfWithdrawn(ctx context.Context, userID uint64) (*float32, error) {
	row := o.db.QueryRowContext(ctx, SelectSumTotalByUserID, userID)

	var sum float32
	if err := row.Scan(&sum); err != nil {
		logger.Log.Error("error sum accrual", zap.Uint64("userId", userID), zap.Error(err))
		return nil, err
	}

	return &sum, nil
}
