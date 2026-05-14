package listing

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

var ErrNotFound = errors.New("listing: not found")

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

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
			Omit(clause.Associations, "FirstSeenAt").
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "board"}, {Name: "mls"}},
				DoUpdates: clause.AssignmentColumns([]string{"latitude", "longitude", "address", "status", "slug"}),
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

const statusAvailable = "1"

func (r *Repository) ListListings(ctx context.Context, where Where, page Page, sort Sort) ([]ListingRow, int64, error) {
	latestPrice := r.db.Model(&PriceHistory{}).
		Select("DISTINCT ON (board, mls) board, mls, price").
		Order("board, mls, observed_at DESC")

	var rows []ListingRow
	var total int64

	q := r.db.WithContext(ctx).Model(&Listing{}).
		Select("listings.*, lp.price AS current_price").
		Joins("LEFT JOIN (?) AS lp ON lp.board = listings.board AND lp.mls = listings.mls", latestPrice)

	if !where.ShowUnavailable {
		q = q.Where("listings.status = ?", statusAvailable)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	col := map[string]string{
		"price":                    "lp.price",
		"first_seen_at":            "listings.first_seen_at",
		"commute_seconds_downtown": "listings.commute_seconds_downtown",
	}[sort.By]

	err := q.Order(col + " " + sort.Dir + " NULLS LAST").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(&rows).Error

	return rows, total, err
}

func (r *Repository) ListPendingCommute(ctx context.Context, refresh bool) ([]PendingCommute, error) {
	q := r.db.WithContext(ctx).Model(&Listing{}).
		Select("board, mls, latitude, longitude")
	if !refresh {
		q = q.Where("commute_seconds_downtown IS NULL")
	}
	var rows []PendingCommute
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) GetPendingCommute(ctx context.Context, board, mls int) (*PendingCommute, error) {
	var row PendingCommute
	err := r.db.WithContext(ctx).Model(&Listing{}).
		Select("board, mls, latitude, longitude").
		Where("board = ? AND mls = ?", board, mls).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.Board == 0 && row.MLS == 0 {
		return nil, ErrNotFound
	}
	return &row, nil
}

func (r *Repository) UpdateCommute(ctx context.Context, board, mls int, seconds int, computedAt time.Time) error {
	res := r.db.WithContext(ctx).Model(&Listing{}).
		Where("board = ? AND mls = ?", board, mls).
		Updates(map[string]any{
			"commute_seconds_downtown": seconds,
			"commute_computed_at":      computedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetListing(ctx context.Context, board, mls int) (*Listing, error) {
	var l Listing
	err := r.db.WithContext(ctx).
		Preload("PriceHistories", func(db *gorm.DB) *gorm.DB {
			return db.Order("observed_at DESC")
		}).
		Where("board = ? AND mls = ?", board, mls).
		First(&l).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &l, err
}
