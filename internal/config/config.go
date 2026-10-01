package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret            string `env:"JWT_SECRET"`
	LogLevel             string `env:"LOG_LEVEL"`
}

func ReadConfig() (Config, error) {
	var config Config

	flag.StringVar(&config.RunAddress, "a", "localhost:8080", "address and port to run gophermart")
	flag.StringVar(&config.DatabaseURI, "d", "", "database url")
	flag.StringVar(&config.AccrualSystemAddress, "r", "localhost:8080", "accrual system address")
	flag.StringVar(&config.JWTSecret, "j", "", "jwt secret for authorization, random when not set")
	flag.StringVar(&config.LogLevel, "l", "info", "log level")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}
