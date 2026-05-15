package refresh

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/transit"
	"github.com/google/uuid"
)

var ErrBusy = errors.New("gtfs refresh: already in progress")

type StopRepo interface {
	Replace(ctx context.Context, stops []transit.Stop) error
}

type RunRepo interface {
	Start(ctx context.Context, scheduleID *uuid.UUID) (*Run, error)
	FinishSuccess(ctx context.Context, id uuid.UUID, stopsWritten int) (time.Time, error)
	FinishError(ctx context.Context, id uuid.UUID, message string) (time.Time, error)
	Get(ctx context.Context, id uuid.UUID) (*Run, error)
	List(ctx context.Context, where Where, page Page) ([]Run, int64, error)
}

type Worker struct {
	stops   StopRepo
	runs    RunRepo
	cfg     config.TransitConfig
	rootCtx context.Context
	wg      sync.WaitGroup
	busy    atomic.Bool
}

func NewWorker(stops StopRepo, runs RunRepo, cfg config.TransitConfig) *Worker {
	return &Worker{stops: stops, runs: runs, cfg: cfg}
}

func (w *Worker) Bind(rootCtx context.Context) { w.rootCtx = rootCtx }
func (w *Worker) Wait()                        { w.wg.Wait() }

func (w *Worker) Start() (uuid.UUID, error) {
	return w.start(nil)
}

func (w *Worker) StartForSchedule(scheduleID uuid.UUID) (uuid.UUID, error) {
	return w.start(&scheduleID)
}

func (w *Worker) GetRun(ctx context.Context, id uuid.UUID) (*Run, error) {
	return w.runs.Get(ctx, id)
}

func (w *Worker) ListRuns(ctx context.Context, where Where, page Page) ([]Run, int64, error) {
	return w.runs.List(ctx, where, page)
}

func (w *Worker) start(scheduleID *uuid.UUID) (uuid.UUID, error) {
	if !w.busy.CompareAndSwap(false, true) {
		return uuid.Nil, ErrBusy
	}
	run, err := w.runs.Start(w.rootCtx, scheduleID)
	if err != nil {
		w.busy.Store(false)
		return uuid.Nil, err
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer w.busy.Store(false)
		w.execute(w.rootCtx, run)
	}()
	return run.ID, nil
}

func (w *Worker) execute(ctx context.Context, run *Run) {
	tmpDir, err := os.MkdirTemp("", "gtfs-refresh-*")
	if err != nil {
		w.finishError(ctx, run.ID, "mkdir temp: "+err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)

	sources := DownloadFeeds(DefaultFeeds, DownloadOptions{CacheDir: tmpDir})

	stops, err := ComputeStops(sources, w.cfg)
	if err != nil {
		w.finishError(ctx, run.ID, err.Error())
		return
	}

	if err := w.stops.Replace(ctx, stops); err != nil {
		w.finishError(ctx, run.ID, "replace transit_stops: "+err.Error())
		return
	}

	if _, err := w.runs.FinishSuccess(ctx, run.ID, len(stops)); err != nil {
		slog.Error("gtfs_refresh_runs finish success update failed", "err", err, "run_id", run.ID)
	}
}

func (w *Worker) finishError(ctx context.Context, runID uuid.UUID, message string) {
	slog.Error("gtfs refresh failed", "run_id", runID, "err", message)
	if _, err := w.runs.FinishError(ctx, runID, message); err != nil {
		slog.Error("gtfs_refresh_runs finish error update failed", "err", err, "run_id", runID)
	}
}
