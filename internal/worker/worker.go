package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/realtor"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scraperun"
	"github.com/google/uuid"
)

// ErrBusy is returned by StartScrape when a scrape is already in flight.
// Only one scrape may run at a time; callers should retry later.
var ErrBusy = errors.New("worker: scrape already in progress")

type RealtorClient interface {
	FetchPrices(ctx context.Context) ([]listing.Observation, error)
}

type ListingRepository interface {
	UpsertListings(ctx context.Context, obs []listing.Observation) (int, error)
}

type ScrapeRunRepository interface {
	Start(ctx context.Context, source string, scheduleID *uuid.UUID) (*scraperun.ScrapeRun, error)
	FinishSuccess(ctx context.Context, id uuid.UUID, totalCount, newCount int) (time.Time, error)
	FinishError(ctx context.Context, id uuid.UUID, kind, message string, totalCount, newCount int) (time.Time, error)
	Get(ctx context.Context, id uuid.UUID) (*scraperun.ScrapeRun, error)
}

type Worker struct {
	realtor RealtorClient
	repo    ListingRepository
	runs    ScrapeRunRepository

	rootCtx context.Context
	wg      sync.WaitGroup
	busy    atomic.Bool
}

func New(rc RealtorClient, repo ListingRepository, runs ScrapeRunRepository) *Worker {
	return &Worker{
		realtor: rc,
		repo:    repo,
		runs:    runs,
	}
}

// Bind associates the worker with a process-lifetime context. Background
// scrapes inherit this context (NOT the HTTP request context that triggered
// them) so that a scrape — which can run for minutes across many pages — is
// not aborted when the HTTP client disconnects after receiving its 202. The
// scrape only stops when the process is shutting down.
func (w *Worker) Bind(rootCtx context.Context) {
	w.rootCtx = rootCtx
}

// StartScrape kicks off a manual scrape in the background and returns its
// run id. Returns ErrBusy if another scrape is already running.
func (w *Worker) StartScrape() (uuid.UUID, error) {
	return w.startScrape(nil)
}

// StartScrapeForSchedule kicks off a scheduled scrape. The resulting
// scrape_runs row is stamped with scheduleID so callers can reconstruct
// run history per schedule.
func (w *Worker) StartScrapeForSchedule(scheduleID uuid.UUID) (uuid.UUID, error) {
	return w.startScrape(&scheduleID)
}

func (w *Worker) startScrape(scheduleID *uuid.UUID) (uuid.UUID, error) {
	if !w.busy.CompareAndSwap(false, true) {
		return uuid.Nil, ErrBusy
	}
	run, err := w.runs.Start(w.rootCtx, scraperun.SourceRealtor, scheduleID)
	if err != nil {
		w.busy.Store(false)
		return uuid.Nil, err
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer w.busy.Store(false)
		w.executeScrape(w.rootCtx, run)
	}()
	return run.ID, nil
}

// Wait blocks until any in-flight scrape finishes. Call after server shutdown.
func (w *Worker) Wait() { w.wg.Wait() }

// GetRun returns the persisted scrape run with the given id.
func (w *Worker) GetRun(ctx context.Context, id uuid.UUID) (*scraperun.ScrapeRun, error) {
	return w.runs.Get(ctx, id)
}

func (w *Worker) executeScrape(ctx context.Context, run *scraperun.ScrapeRun) {
	observations, fetchErr := w.realtor.FetchPrices(ctx)
	totalCount := len(observations)
	newCount := 0

	if fetchErr != nil {
		kind := classifyError(fetchErr)
		if _, updateErr := w.runs.FinishError(ctx, run.ID, kind, fetchErr.Error(), totalCount, newCount); updateErr != nil {
			slog.Error("scrape_runs finish error update failed", "err", updateErr, "run_id", run.ID)
		}
		return
	}

	inserted, upsertErr := w.repo.UpsertListings(ctx, observations)
	if upsertErr != nil {
		slog.Error("upsert listings failed", "err", upsertErr)
		if _, updateErr := w.runs.FinishError(ctx, run.ID, scraperun.ErrorKindUnknown, upsertErr.Error(), totalCount, newCount); updateErr != nil {
			slog.Error("scrape_runs finish error update failed", "err", updateErr, "run_id", run.ID)
		}
		return
	}
	newCount = inserted

	if _, updateErr := w.runs.FinishSuccess(ctx, run.ID, totalCount, newCount); updateErr != nil {
		slog.Error("scrape_runs finish success update failed", "err", updateErr, "run_id", run.ID)
	}
}

func classifyError(err error) string {
	switch {
	case errors.Is(err, realtor.ErrForbidden):
		return scraperun.ErrorKindForbidden
	case errors.Is(err, realtor.ErrNullCount):
		return scraperun.ErrorKindNullCount
	default:
		return scraperun.ErrorKindUnknown
	}
}
