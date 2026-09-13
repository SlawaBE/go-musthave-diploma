package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/SlawaBE/go-musthave-diploma/internal/config"
	"github.com/SlawaBE/go-musthave-diploma/internal/db"
	"github.com/SlawaBE/go-musthave-diploma/internal/handler"
	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/middleware"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
)

func Run(config config.Config) {
	logger.Initialize(config.LogLevel)

	database := initDatabase(config)
	defer database.Close()

	var r http.Handler = InitRouter(database, config)

	r = middleware.GZip(r)
	r = middleware.RequestLogger(r)

	err := http.ListenAndServe(config.RunAddress, r)
	if err != nil {
		log.Fatal(err)
	}
}

func initDatabase(config config.Config) *sql.DB {
	database, err := db.NewDB(config.DatabaseURI)
	if err != nil {
		logger.Log.Error("error open db connect", logger.Err(err))
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = database.PingContext(ctx)
	if err != nil {
		logger.Log.Error("Error ping database", logger.Err(err))
		os.Exit(3)
	}

	err = db.RunMigrations(database, config.DatabaseURI)
	if err != nil {
		logger.Log.Error("Error migration", logger.Err(err))
		os.Exit(4)
	}
	return database
}

func InitRouter(database *sql.DB, config config.Config) *http.ServeMux {
	r := http.NewServeMux()

	if config.JWTSecret == "" {
		config.JWTSecret = rand.Text()
		logger.Log.Warn("JWT secret is not set. generate random")
		logger.Log.Debug("JWT secret: " + config.JWTSecret)
	}

	ts := service.NewTokenService(config.JWTSecret, time.Minute*30)
	ur := repository.NewUserRepository(database)
	or := repository.NewOrderRepository(database)
	wr := repository.NewWithdrawRepository(database)

	as := service.NewAccrualService(config.AccrualSystemAddress, or)
	as.Run(context.Background())

	registerHandler := handler.NewRegisterHandler(ur, ts)
	loginHandler := handler.NewLoginHandler(ur, ts)
	ordersUploadHandler := handler.NewOrdersUploadHandler(or, as)
	ordersListHandler := handler.NewOrdersListHandler(or)
	balanceHandler := handler.NewBalanceHandler(or, wr)
	withdrawUploadHandler := handler.NewWithdrawUploadHandler(wr, or)
	withdrawListHandler := handler.NewWithdrawListHandler(wr)

	authMiddleware := ts.CreateAuthMiddleware()

	r.Handle("/api/user/register", registerHandler)
	r.Handle("/api/user/login", loginHandler)

	r.Handle("POST /api/user/orders", authMiddleware(ordersUploadHandler))
	r.Handle("GET /api/user/orders", authMiddleware(ordersListHandler))
	r.Handle("GET /api/user/balance", authMiddleware(balanceHandler))
	r.Handle("POST /api/user/balance/withdraw", authMiddleware(withdrawUploadHandler))
	r.Handle("GET /api/user/withdrawals", authMiddleware(withdrawListHandler))

	return r
}
