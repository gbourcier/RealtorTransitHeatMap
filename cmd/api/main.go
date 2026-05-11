package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/api"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/db"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/realtor"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/schedule"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scraperun"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/worker"
	"github.com/joho/godotenv"
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

	gormDB, err := db.Open(cfg.DB.URL)
	if err != nil {
		return err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	realtorClient := realtor.NewClient(cfg.Realtor)
	listings := listing.NewRepository(gormDB)
	runs := scraperun.NewRepository(gormDB)
	schedules := schedule.NewRepository(gormDB)

	w := worker.New(realtorClient, listings, runs)
	w.Bind(ctx)

	scheduler := schedule.New(schedules, w)
	if err := scheduler.Start(ctx); err != nil {
		return err
	}

	srv := api.NewServer(cfg.HTTP.Addr, w, schedules, scheduler, listings)

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", cfg.HTTP.Addr)
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
	scheduler.Stop()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Warn("server shutdown error", "err", err)
	}
	w.Wait()
	return nil
}
