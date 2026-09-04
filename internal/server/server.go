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
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func Run(config config.Config) {
	logger.Initialize("info")

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
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = database.PingContext(ctx)
	if err != nil {
		logger.Log.Fatal("Error ping database", zap.Error(err))
		os.Exit(3)
	}

	err = db.RunMigrations(database, config.DatabaseURI)
	if err != nil {
		logger.Log.Fatal("Error migration", zap.Error(err))
		os.Exit(4)
	}
	return database
}

func InitRouter(database *sql.DB, config config.Config) chi.Router {
	r := chi.NewRouter()

	jwtSecret := rand.Text()
	logger.Log.Info("JWT secret: " + jwtSecret) //TODO change to debug or delete
	ts := service.NewTokenService(jwtSecret, time.Minute*30)
	as := service.NewAccrualService(config.AccrualSystemAddress)
	ur := repository.NewUserRepository(database)
	or := repository.NewOrderRepository(database)
	wr := repository.NewWitdrawRepository(database)

	registerHandler := handler.NewRegisterHandler(ur, ts)
	loginHandler := handler.NewLoginHandler(ur, ts)
	ordersUploadHandler := handler.NewOrdersUploadHandler(or, ur, as)
	ordersListHandler := handler.NewOrdersListHandler(or, ur)
	balanceHandler := handler.NewBalanceHandler(or, ur, wr)
	withdrawUploadHandler := handler.NewWitdrawUploadHandler(wr, ur, or)
	withdrawListHandler := handler.NewWithdrawListHandler(wr, ur)

	authMiddleware := ts.CreateAuthMiddleware()

	r.Route("/api/user", func(r chi.Router) {
		r.Handle("/register", registerHandler)
		r.Handle("/login", loginHandler)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/orders", ordersUploadHandler.ServeHTTP)
			r.Get("/orders", ordersListHandler.ServeHTTP)
			r.Get("/balance", balanceHandler.ServeHTTP)
			r.Post("/balance/withdraw", withdrawUploadHandler.ServeHTTP)
			r.Get("/withdrawals", withdrawListHandler.ServeHTTP)
		})
	})

	return r
}
