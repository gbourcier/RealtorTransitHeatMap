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
	List(ctx context.Context, where scraperun.Where, page scraperun.Page) ([]scraperun.ScrapeRun, int64, error)
}

type TransitTrigger interface {
	StartComputation(refresh bool) error
}

type Worker struct {
	realtor RealtorClient
	repo    ListingRepository
	runs    ScrapeRunRepository
	transit TransitTrigger
	rootCtx context.Context
	wg      sync.WaitGroup
	busy    atomic.Bool
}

func New(rc RealtorClient, repo ListingRepository, runs ScrapeRunRepository, transit TransitTrigger) *Worker {
	return &Worker{
		realtor: rc,
		repo:    repo,
		runs:    runs,
		transit: transit,
	}
}

func (w *Worker) Bind(rootCtx context.Context) {
	w.rootCtx = rootCtx
}

func (w *Worker) StartScrape() (uuid.UUID, error) {
	return w.startScrape(nil)
}

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

func (w *Worker) Wait() { w.wg.Wait() }

func (w *Worker) GetRun(ctx context.Context, id uuid.UUID) (*scraperun.ScrapeRun, error) {
	return w.runs.Get(ctx, id)
}

func (w *Worker) ListRuns(ctx context.Context, where scraperun.Where, page scraperun.Page) ([]scraperun.ScrapeRun, int64, error) {
	return w.runs.List(ctx, where, page)
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

	if w.transit != nil {
		if err := w.transit.StartComputation(false); err != nil {
			slog.Info("transit trigger skipped", "reason", err)
		}
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
