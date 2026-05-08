package listing

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) UpsertListings(ctx context.Context, listings []Listing) error {
	// TODO: implement upsert into listings + listing_price_history
	return nil
}
