package repository

import (
	"context"
	"database/sql"
	"encoding/hex"

	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/model"
	"github.com/SlawaBE/go-musthave-diploma/internal/utils/hash"
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
	INSERT_USER = `INSERT INTO users (login, password_hash) VALUES ($1, $2);`
	SELECT_USER = `SELECT login, password_hash FROM users WHERE login = $1;`
)

func (u *UserRepository) CreateUser(ctx context.Context, user model.User) error {
	tx, err := u.db.Begin()
	if err != nil {
		logger.Log.Error("error begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, INSERT_USER)
	if err != nil {
		logger.Log.Error("error prepare statement", zap.Error(err))
		return err
	}
	defer stmt.Close()

	user.Password = hex.EncodeToString(hash.Sha256([]byte(user.Password)))

	_, err = stmt.ExecContext(ctx, user.Login, user.Password)
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

func (u *UserRepository) GetUser(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	rows := u.db.QueryRowContext(ctx, SELECT_USER, login)
	var err error

	if err = rows.Scan(&user.Login, &user.Password); err != nil {
		logger.Log.Error("error get login", zap.String("login", login), zap.Error(err))
		return nil, err
	}
	return &user, nil
}
