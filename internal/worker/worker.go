package worker

import (
	"context"
	"log/slog"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

const queueSize = 32

type RealtorClient interface {
	FetchPrices(ctx context.Context, c listing.FetchCriteria) ([]listing.Listing, error)
}

type Repository interface {
	UpsertListings(ctx context.Context, l []listing.Listing) error
}

type Worker struct {
	realtor RealtorClient
	repo    Repository
	jobs    chan job
}

func New(rc RealtorClient, repo Repository) *Worker {
	return &Worker{
		realtor: rc,
		repo:    repo,
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
			listings, err := w.realtor.FetchPrices(ctx, j.criteria)
			if err == nil {
				if upsertErr := w.repo.UpsertListings(ctx, listings); upsertErr != nil {
					slog.Error("upsert listings failed", "err", upsertErr)
				}
			}
			j.reply <- result{listings: listings, err: err}
		}
	}
}

func (w *Worker) FetchPrices(ctx context.Context, c listing.FetchCriteria) ([]listing.Listing, error) {
	reply := make(chan result, 1)
	select {
	case w.jobs <- job{criteria: c, reply: reply}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case r := <-reply:
		return r.listings, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
