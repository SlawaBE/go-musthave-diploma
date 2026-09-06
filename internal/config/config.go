package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddress   			string `env:"RUN_ADDRESS"`
	DatabaseURI     		string `env:"DATABASE_URI"`
	AccrualSystemAddress    string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func ReadConfig() (Config, error) {
	var config Config

	flag.StringVar(&config.RunAddress, "a", "localhost:8080", "address and port to run gophermart")
	flag.StringVar(&config.DatabaseURI, "d", "", "database url")
	flag.StringVar(&config.AccrualSystemAddress, "r", "localhost:8080", "accrual system address")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}
