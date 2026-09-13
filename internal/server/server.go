package server

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SlawaBE/go-musthave-diploma/internal/config"
	"github.com/SlawaBE/go-musthave-diploma/internal/db"
	"github.com/SlawaBE/go-musthave-diploma/internal/handler"
	"github.com/SlawaBE/go-musthave-diploma/internal/logger"
	"github.com/SlawaBE/go-musthave-diploma/internal/middleware"
	"github.com/SlawaBE/go-musthave-diploma/internal/repository"
	"github.com/SlawaBE/go-musthave-diploma/internal/service"
)

type App struct {
	httpServer     *http.Server
	accrualService *service.AccrualService
}

func NewApp(config.Config) *App {
	return &App{}
}

func (a *App) Run(config config.Config) {
	logger.Initialize(config.LogLevel)
	logger.Log.Info("Starting server")
	logger.Log.Info("Logger has been initialized")

	database := a.initDatabase(config)
	defer database.Close()
	logger.Log.Info("Database has been initialized")

	var r http.Handler = a.initRouter(database, config)
	r = middleware.GZip(r)
	r = middleware.RequestLogger(r)
	logger.Log.Info("Router has been initialized")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a.httpServer = &http.Server{
		Addr:    config.RunAddress,
		Handler: r,
	}

	a.accrualService.Run(ctx)

	go func() {
		logger.Log.Info("Server started")
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("Server error", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Log.Info("Graceful shutdown")

	stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Server shutdown error", logger.Err(err))
	}

	a.accrualService.Stop()

	logger.Log.Info("Server has been stopped")
}

func (a *App) initDatabase(config config.Config) *sql.DB {
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

func (a *App) initRouter(database *sql.DB, config config.Config) *http.ServeMux {
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
	a.accrualService = as

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
