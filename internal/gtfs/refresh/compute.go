package refresh

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/transit"
)

type DownloadOptions struct {
	CacheDir     string
	SkipDownload bool
}

func DownloadFeeds(feeds []FeedDef, opts DownloadOptions) []gtfs.FeedSource {
	sources := make([]gtfs.FeedSource, 0, len(feeds))
	for _, fd := range feeds {
		zipPath := filepath.Join(opts.CacheDir, fd.Agency+".zip")
		if !opts.SkipDownload || !fileExists(zipPath) {
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
	return sources
}

func ComputeStops(sources []gtfs.FeedSource, cfg config.TransitConfig) ([]transit.Stop, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("no GTFS feeds available")
	}

	slog.Info("loading GTFS dataset",
		"feeds", len(sources),
		"snapshot", cfg.SnapshotKey,
		"day", cfg.SnapshotDay.String(),
		"hour", cfg.SnapshotHour,
	)

	t0 := time.Now()
	dataset, err := gtfs.Load(sources, cfg.WalkSpeedMps, cfg.WalkDetour, cfg.SnapshotDay, cfg.SnapshotHour)
	if err != nil {
		return nil, fmt.Errorf("load gtfs: %w", err)
	}
	slog.Info("dataset loaded",
		"stops", dataset.StopCount(),
		"connections", dataset.ConnectionCount(),
		"elapsed", time.Since(t0),
	)

	targets, err := dataset.ResolveTarget(cfg.ReferenceLat, cfg.ReferenceLon, "McGill")
	if err != nil {
		return nil, fmt.Errorf("resolve target: %w", err)
	}
	slog.Info("resolved reference targets", "count", len(targets), "key", cfg.ReferenceKey)

	arriveBy := cfg.SnapshotHour * 3600
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

	now := time.Now()
	stops := make([]transit.Stop, 0, len(results))
	for _, r := range results {
		stops = append(stops, transit.Stop{
			StopID:                 gtfs.PrefixedStopID(r.Agency, r.StopID),
			Agency:                 r.Agency,
			Name:                   r.Name,
			Latitude:               r.Latitude,
			Longitude:              r.Longitude,
			CommuteSecondsToMcGill: r.CommuteSecondsToTarget,
			SnapshotKey:            cfg.SnapshotKey,
			ComputedAt:             now,
		})
	}
	return stops, nil
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
	req.Header.Set("User-Agent", "RealtorTransitHeatMap-gtfs-refresh/1.0")
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
