package main

import (
	"fmt"
	"os"

	"github.com/SlawaBE/go-musthave-diploma/internal/config"
	"github.com/SlawaBE/go-musthave-diploma/internal/server"
)

func main() {
	conf, err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error reading configuration")
		os.Exit(1)
	}
	server.Run(conf)
}
