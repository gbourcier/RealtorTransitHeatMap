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

func (r *Repository) UpsertListings(ctx context.Context, listings []Listing) error {
	if len(listings) == 0 {
		return nil
	}

	observedAt := time.Now()
	prices := make([]PriceHistory, len(listings))
	for i, l := range listings {
		prices[i] = PriceHistory{
			Board:      l.Board,
			MLS:        l.MLS,
			ObservedAt: observedAt,
			Price:      l.Price,
		}
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
}

func (r *Repository) LatestPrice(ctx context.Context, board, mls string) (PriceHistory, error) {
	var ph PriceHistory
	err := r.db.WithContext(ctx).
		Where("board = ? AND mls = ?", board, mls).
		Order("observed_at DESC").
		First(&ph).Error
	return ph, err
}
