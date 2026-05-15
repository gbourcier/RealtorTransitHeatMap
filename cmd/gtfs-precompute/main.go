package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/db"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs"
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

	sources := make([]gtfs.FeedSource, 0, len(defaultFeeds))
	for _, fd := range defaultFeeds {
		zipPath := filepath.Join(cacheDir, fd.Agency+".zip")
		if !skipDownload || !fileExists(zipPath) {
			slog.Info("downloading feed", "agency", fd.Agency, "url", fd.URL)
			if err := downloadFile(fd.URL, zipPath); err != nil {
				slog.Warn("download failed; skipping feed", "agency", fd.Agency, "err", err)
				continue
			}
		}
		if !fileExists(zipPath) {
			continue
		}
		sources = append(sources, fd.toSource(zipPath))
	}

	if len(sources) == 0 {
		return fmt.Errorf("no GTFS feeds available")
	}

	slog.Info("loading GTFS dataset",
		"feeds", len(sources),
		"snapshot", cfg.Transit.SnapshotKey,
		"day", cfg.Transit.SnapshotDay.String(),
		"hour", cfg.Transit.SnapshotHour,
	)

	t0 := time.Now()
	dataset, err := gtfs.Load(sources, cfg.Transit.WalkSpeedMps, cfg.Transit.WalkDetour, cfg.Transit.SnapshotDay, cfg.Transit.SnapshotHour)
	if err != nil {
		return fmt.Errorf("load gtfs: %w", err)
	}
	slog.Info("dataset loaded",
		"stops", dataset.StopCount(),
		"connections", dataset.ConnectionCount(),
		"elapsed", time.Since(t0),
	)

	targets, err := dataset.ResolveTarget(cfg.Transit.ReferenceLat, cfg.Transit.ReferenceLon, "McGill")
	if err != nil {
		return fmt.Errorf("resolve target: %w", err)
	}
	slog.Info("resolved reference targets", "count", len(targets), "key", cfg.Transit.ReferenceKey)

	arriveBy := cfg.Transit.SnapshotHour * 3600
	t1 := time.Now()
	results := dataset.ComputeBackward(targets, arriveBy)
	slog.Info("CSA complete", "elapsed", time.Since(t1))

	reachable := 0
	for _, r := range results {
		if r.CommuteSecondsToTarget != nil {
			reachable++
		}
	}
	slog.Info("computed commute", "reachable", reachable, "total", len(results))

	if dryRun {
		slog.Info("dry-run: skipping database write")
		return nil
	}

	stops := make([]transit.Stop, 0, len(results))
	now := time.Now()
	for _, r := range results {
		stops = append(stops, transit.Stop{
			StopID:                 gtfs.PrefixedStopID(r.Agency, r.StopID),
			Agency:                 r.Agency,
			Name:                   r.Name,
			Latitude:               r.Latitude,
			Longitude:              r.Longitude,
			CommuteSecondsToMcGill: r.CommuteSecondsToTarget,
			SnapshotKey:            cfg.Transit.SnapshotKey,
			ComputedAt:             now,
		})
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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func downloadFile(url, dest string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "RealtorTransitHeatMap-gtfs-precompute/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
