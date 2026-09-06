package repository

import (
	"context"
	"database/sql"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"go.uber.org/zap"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

const (
	InsertUser        = `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`
	SelectUserByLogin = `SELECT id, login, password_hash FROM users WHERE login = $1;`
)

func (u *UserRepository) SaveUser(ctx context.Context, user *model.User) error {
	tx, err := u.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, InsertUser)
	if err != nil {
		logger.Log.Error("error prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, user.Login, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		logger.Log.Error("error exec statement", zap.Error(err))
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Error("error commit transaction", zap.Error(err))
		return err
	}

	logger.Log.Info("user saved", zap.Uint64("id", user.ID))
	return nil
}

func (u *UserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	rows := u.db.QueryRowContext(ctx, SelectUserByLogin, login)
	var err error

	if err = rows.Scan(&user.ID, &user.Login, &user.PasswordHash); err != nil {
		logger.Log.Error("error get login", zap.String("login", login), zap.Error(err))
		return nil, err
	}
	return &user, nil
}
