package worker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scraperun"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/worker/realtor"
	"github.com/google/uuid"
)

const queueSize = 32

type RealtorClient interface {
	FetchPrices(ctx context.Context) ([]listing.Listing, error)
}

type ListingRepository interface {
	UpsertListings(ctx context.Context, l []listing.Listing) (int, error)
}

type ScrapeRunRepository interface {
	Start(ctx context.Context, source string) (*scraperun.ScrapeRun, error)
	FinishSuccess(ctx context.Context, id uuid.UUID, totalCount, newCount int) (time.Time, error)
	FinishError(ctx context.Context, id uuid.UUID, kind, message string, totalCount, newCount int) (time.Time, error)
}

type Worker struct {
	realtor RealtorClient
	repo    ListingRepository
	runs    ScrapeRunRepository
	jobs    chan job
}

func New(rc RealtorClient, repo ListingRepository, runs ScrapeRunRepository) *Worker {
	return &Worker{
		realtor: rc,
		repo:    repo,
		runs:    runs,
		jobs:    make(chan job, queueSize),
	}
}

func (w *Worker) Run(ctx context.Context) {
	slog.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			slog.Info("worker shutting down")
			return
		case j := <-w.jobs:
			j.reply <- w.executeScrape(ctx)
		}
	}
}

// ExecuteScrape submits a scrape job to the worker and waits for the result.
// It is the public entry point invoked from the HTTP handler.
func (w *Worker) ExecuteScrape(ctx context.Context) (ScrapeResult, error) {
	reply := make(chan ScrapeResult, 1)
	select {
	case w.jobs <- job{reply: reply}:
	case <-ctx.Done():
		return ScrapeResult{}, ctx.Err()
	}

	select {
	case r := <-reply:
		return r, nil
	case <-ctx.Done():
		return ScrapeResult{}, ctx.Err()
	}
}

func (w *Worker) executeScrape(ctx context.Context) ScrapeResult {
	run, err := w.runs.Start(ctx, scraperun.SourceRealtor)
	if err != nil {
		slog.Error("scrape_runs start failed", "err", err)
		// We can't persist a row, but still return a meaningful result.
		now := time.Now()
		return ScrapeResult{
			Status:      scraperun.StatusError,
			StartedAt:   now,
			CompletedAt: now,
			ErrorKind:   scraperun.ErrorKindUnknown,
			Error:       err.Error(),
		}
	}

	listings, fetchErr := w.realtor.FetchPrices(ctx)
	totalCount := len(listings)
	newCount := 0

	if fetchErr != nil {
		kind := classifyError(fetchErr)
		completedAt, updateErr := w.runs.FinishError(ctx, run.ID, kind, fetchErr.Error(), totalCount, newCount)
		if updateErr != nil {
			slog.Error("scrape_runs finish error update failed", "err", updateErr, "run_id", run.ID)
		}
		return ScrapeResult{
			Status:      scraperun.StatusError,
			StartedAt:   run.StartedAt,
			CompletedAt: completedAt,
			TotalCount:  totalCount,
			NewCount:    newCount,
			ErrorKind:   kind,
			Error:       fetchErr.Error(),
		}
	}

	inserted, upsertErr := w.repo.UpsertListings(ctx, listings)
	if upsertErr != nil {
		slog.Error("upsert listings failed", "err", upsertErr)
		completedAt, updateErr := w.runs.FinishError(ctx, run.ID, scraperun.ErrorKindUnknown, upsertErr.Error(), totalCount, newCount)
		if updateErr != nil {
			slog.Error("scrape_runs finish error update failed", "err", updateErr, "run_id", run.ID)
		}
		return ScrapeResult{
			Status:      scraperun.StatusError,
			StartedAt:   run.StartedAt,
			CompletedAt: completedAt,
			TotalCount:  totalCount,
			NewCount:    newCount,
			ErrorKind:   scraperun.ErrorKindUnknown,
			Error:       upsertErr.Error(),
		}
	}
	newCount = inserted

	completedAt, updateErr := w.runs.FinishSuccess(ctx, run.ID, totalCount, newCount)
	if updateErr != nil {
		slog.Error("scrape_runs finish success update failed", "err", updateErr, "run_id", run.ID)
	}
	return ScrapeResult{
		Status:      scraperun.StatusSuccess,
		StartedAt:   run.StartedAt,
		CompletedAt: completedAt,
		TotalCount:  totalCount,
		NewCount:    newCount,
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
