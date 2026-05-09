package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL    string
	HTTPAddr       string
	RealtorBaseURL string
	MockRealtorAPI bool
	PolygonWKT     string
	PriceMin       string
	PriceMax       string
	BedRange       string
	BathRange      string
	LatitudeMax    string
	LatitudeMin    string
	LongitudeMax   string
	LongitudeMin   string
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

	polygonWKT := os.Getenv("POLYGON_WKT")
	if polygonWKT == "" {
		return nil, errors.New("POLYGON_WKT is required")
	}

	realtorBaseURL := os.Getenv("REALTOR_BASE_URL")
	if realtorBaseURL == "" {
		realtorBaseURL = "https://api2.realtor.ca"
	}

	sslMode := ""
	if os.Getenv("POSTGRES_ENABLE_SSL") == "false" {
		sslMode = "?sslmode=disable"
	}

	return &Config{
		DatabaseURL:    fmt.Sprintf("postgres://%s:%s@%s:%s/%s%s", user, password, host, port, db, sslMode),
		HTTPAddr:       addr,
		RealtorBaseURL: realtorBaseURL,
		MockRealtorAPI: os.Getenv("MOCK_REALTOR_API") == "true",
		PolygonWKT:     polygonWKT,
		PriceMin:       os.Getenv("PRICE_MIN"),
		PriceMax:       os.Getenv("PRICE_MAX"),
		BedRange:       os.Getenv("BED_RANGE"),
		BathRange:      os.Getenv("BATH_RANGE"),
		LatitudeMax:    os.Getenv("LAT_MAX"),
		LatitudeMin:    os.Getenv("LAT_MIN"),
		LongitudeMax:   os.Getenv("LON_MAX"),
		LongitudeMin:   os.Getenv("LON_MIN"),
	}, nil
}
