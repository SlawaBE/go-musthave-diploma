package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
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
	insertUser        = `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`
	selectUserByLogin = `SELECT id, login, password_hash FROM users WHERE login = $1;`
)

func (u *UserRepository) SaveUser(ctx context.Context, user *model.User) error {
	tx, err := u.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", logger.Err(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, insertUser)
	if err != nil {
		logger.Log.Error("error prepare statement", logger.Err(err))
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, user.Login, user.PasswordHash).Scan(&user.ID)
	if err != nil {
		logger.Log.Error("error exec statement", logger.Err(err))
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Log.Error("error commit transaction", logger.Err(err))
		return err
	}

	logger.Log.Info("user saved", slog.Uint64("id", user.ID))
	return nil
}

func (u *UserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	rows := u.db.QueryRowContext(ctx, selectUserByLogin, login)
	var err error

	if err = rows.Scan(&user.ID, &user.Login, &user.PasswordHash); err != nil {
		logger.Log.Error("error get login", slog.String("login", login), logger.Err(err))
		return nil, err
	}
	return &user, nil
}
