package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
}

func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:3000"
	}
	return &Config{
		DatabaseURL: dbURL,
		HTTPAddr:    addr,
	}, nil
}
