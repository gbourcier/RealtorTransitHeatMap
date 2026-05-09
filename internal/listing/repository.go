package listing

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// UpsertListings persists the given observations (one listing row + one price
// history row each) and returns the count of newly inserted listings — i.e.
// listings whose (board, mls) key did not already exist. Updates to existing
// rows are not counted toward inserted.
func (r *Repository) UpsertListings(ctx context.Context, obs []Observation) (inserted int, err error) {
	if len(obs) == 0 {
		return 0, nil
	}

	observedAt := time.Now()
	listings := make([]Listing, len(obs))
	prices := make([]PriceHistory, len(obs))
	for i, o := range obs {
		listings[i] = o.Listing
		prices[i] = PriceHistory{
			Board:      o.Listing.Board,
			MLS:        o.Listing.MLS,
			ObservedAt: observedAt,
			Price:      o.Price,
		}
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		keys := make([][]any, len(listings))
		for i, l := range listings {
			keys[i] = []any{l.Board, l.MLS}
		}
		var existing int64
		if err := tx.Model(&Listing{}).
			Where("(board, mls) IN ?", keys).
			Count(&existing).Error; err != nil {
			return err
		}
		inserted = len(listings) - int(existing)

		if err := tx.
			Omit(clause.Associations).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "board"}, {Name: "mls"}},
				DoUpdates: clause.AssignmentColumns([]string{"latitude", "longitude", "address", "status"}),
			}).
			CreateInBatches(listings, 200).Error; err != nil {
			return err
		}
		return tx.
			Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(prices, 200).Error
	})
	if err != nil {
		return 0, err
	}
	return inserted, nil
}

func (r *Repository) LatestPrice(ctx context.Context, board, mls string) (PriceHistory, error) {
	var ph PriceHistory
	err := r.db.WithContext(ctx).
		Where("board = ? AND mls = ?", board, mls).
		Order("observed_at DESC").
		First(&ph).Error
	return ph, err
}
