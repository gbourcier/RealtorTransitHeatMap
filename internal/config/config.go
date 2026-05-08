package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
}

func Load() (*Config, error) {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return nil, errors.New("POSTGRES_USER is required")
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return nil, errors.New("POSTGRES_PASSWORD is required")
	}
	db := os.Getenv("POSTGRES_DB")
	if db == "" {
		return nil, errors.New("POSTGRES_DB is required")
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:3000"
	}

	return &Config{
		DatabaseURL: fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, db),
		HTTPAddr:    addr,
	}, nil
}
