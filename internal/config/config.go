package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", ":3333", "Server address:port")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", ":8080", "Accrual system address")

	flag.Parse()

	if envServerAddress := os.Getenv("RUN_ADDRESS"); envServerAddress != "" {
		cfg.RunAddress = envServerAddress
	}

	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}

	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envAccrualSystemAddress
	}

	return cfg
}
