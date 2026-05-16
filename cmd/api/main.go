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
	"github.com/gbourcier/RealtorTransitHeatMap/internal/dispatch"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs/refresh"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/realtor"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/schedule"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scrape"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scraperun"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/transit"
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
	scrapeRuns := scraperun.NewRepository(gormDB)
	refreshRuns := refresh.NewRepository(gormDB)
	schedules := schedule.NewRepository(gormDB)
	stops := transit.NewRepository(gormDB)

	transitWorker := transit.NewWorker(stops, listings, transit.Config{
		NearestStops: cfg.Transit.NearestStops,
		WalkSpeedMps: cfg.Transit.WalkSpeedMps,
		WalkDetour:   cfg.Transit.WalkDetour,
	})
	transitWorker.Bind(ctx)

	scrapeWorker := scrape.New(realtorClient, listings, scrapeRuns, transitWorker)
	scrapeWorker.Bind(ctx)

	refreshWorker := refresh.NewWorker(stops, refreshRuns, transitWorker, cfg.Transit)
	refreshWorker.Bind(ctx)

	dispatcher := dispatch.New(scrapeWorker, refreshWorker)

	scheduler := schedule.New(schedules, dispatcher)
	if err := scheduler.Start(ctx); err != nil {
		return err
	}

	srv := api.NewServer(cfg.HTTP.Addr, scrapeWorker, refreshWorker, schedules, scheduler, listings, transitWorker, stops)

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
	scrapeWorker.Wait()
	refreshWorker.Wait()
	transitWorker.Wait()
	return nil
}
