package listing

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) UpsertListings(ctx context.Context, listings []Listing) error {
	if len(listings) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, l := range listings {
		batch.Queue(`
			INSERT INTO listings (board, mls, latitude, longitude, address, status)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (board, mls) DO UPDATE SET
				latitude = EXCLUDED.latitude,
				longitude = EXCLUDED.longitude,
				address = EXCLUDED.address,
				status = EXCLUDED.status`,
			l.Board, l.MLS, l.Latitude, l.Longitude, l.Address, l.Status)

		batch.Queue(`
			INSERT INTO listing_price_history (board, mls, observed_at, price)
			VALUES ($1, $2, now(), $3)
			ON CONFLICT (board, mls, observed_at) DO NOTHING`,
			l.Board, l.MLS, l.Price)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch exec[%d]: %w", i, err)
		}
	}
	return nil
}
