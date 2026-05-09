// @title           Realtor Transit Heat Map API
// @version         1.0
// @description     API for fetching real estate listings using the operator-configured search parameters.
// @host            localhost:3000
// @BasePath        /

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/db"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/httpapi"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/worker"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/worker/realtor"
	"github.com/joho/godotenv"

	_ "github.com/gbourcier/RealtorTransitHeatMap/docs"
)

func main() {
	_ = godotenv.Load()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	gormDB, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	realtorClient := realtor.NewClient(cfg)
	repo := listing.NewRepository(gormDB)

	w := worker.New(realtorClient, repo)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()

	srv := httpapi.NewServer(cfg.HTTPAddr, w)

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", cfg.HTTPAddr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("server shutdown error", "err", err)
	}
	wg.Wait()
	return nil
}
