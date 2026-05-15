package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/db"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs/refresh"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/transit"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cacheDir := flag.String("cache", "data/gtfs", "directory to cache downloaded GTFS zips")
	skipDownload := flag.Bool("skip-download", false, "skip downloading if cache files exist")
	dryRun := flag.Bool("dry-run", false, "compute but don't write to database")
	ifEmpty := flag.Bool("if-empty", false, "skip work if transit_stops already has rows")
	flag.Parse()

	if err := run(*cacheDir, *skipDownload, *dryRun, *ifEmpty); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(cacheDir string, skipDownload, dryRun, ifEmpty bool) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if ifEmpty {
		gormDB, err := db.Open(cfg.DB.URL)
		if err != nil {
			return err
		}
		sqlDB, err := gormDB.DB()
		if err != nil {
			return err
		}
		n, err := transit.NewRepository(gormDB).Count(context.Background())
		sqlDB.Close()
		if err != nil {
			return fmt.Errorf("count transit_stops: %w", err)
		}
		if n > 0 {
			slog.Info("transit_stops already populated; skipping", "rows", n)
			return nil
		}
	}

	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("mkdir cache: %w", err)
	}

	sources := refresh.DownloadFeeds(refresh.DefaultFeeds, refresh.DownloadOptions{
		CacheDir:     cacheDir,
		SkipDownload: skipDownload,
	})

	stops, err := refresh.ComputeStops(sources, cfg.Transit)
	if err != nil {
		return err
	}

	if dryRun {
		slog.Info("dry-run: skipping database write", "stops", len(stops))
		return nil
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

	repo := transit.NewRepository(gormDB)
	if err := repo.Replace(context.Background(), stops); err != nil {
		return fmt.Errorf("replace transit_stops: %w", err)
	}
	slog.Info("transit_stops written", "rows", len(stops))
	return nil
}
