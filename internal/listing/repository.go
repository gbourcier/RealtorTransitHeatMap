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
				DoUpdates: clause.AssignmentColumns([]string{"latitude", "longitude", "address", "is_available", "slug", "photo_url", "building_type", "bedroom_count", "bathroom_count", "interior_area_sqft"}),
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

func (r *Repository) ListListings(ctx context.Context, where Where, page Page, sort Sort) ([]ListingRow, int64, error) {
	latestPrice := r.db.Model(&PriceHistory{}).
		Select("DISTINCT ON (board, mls) board, mls, price").
		Order("board, mls, observed_at DESC")

	var rows []ListingRow
	var total int64

	q := r.db.WithContext(ctx).Model(&Listing{}).
		Select("listings.*, lp.price AS current_price, (fav.user_id IS NOT NULL) AS is_favorite").
		Joins("LEFT JOIN (?) AS lp ON lp.board = listings.board AND lp.mls = listings.mls", latestPrice).
		Joins("LEFT JOIN favorites fav ON fav.board = listings.board AND fav.mls = listings.mls AND fav.user_id = ?", where.UserID)

	if !where.ShowUnavailable {
		q = q.Where("listings.is_available = ?", true)
	}
	if where.FavoritesOnly {
		q = q.Where("fav.user_id IS NOT NULL")
	}
	if len(where.BuildingTypes) > 0 {
		q = q.Where("listings.building_type IN ?", where.BuildingTypes)
	}
	if where.MaxPrice != nil {
		q = q.Where("lp.price IS NOT NULL AND lp.price <= ?", *where.MaxPrice)
	}
	if where.MaxCommuteSec != nil {
		q = q.Where("listings.commute_seconds_downtown IS NOT NULL AND listings.commute_seconds_downtown <= ?", *where.MaxCommuteSec)
	}
	if where.NewSince != nil {
		q = q.Where("listings.first_seen_at >= ?", *where.NewSince)
	}
	if where.MinBedrooms != nil {
		q = q.Where("listings.bedroom_count >= ?", *where.MinBedrooms)
	}
	if where.MinBathrooms != nil {
		q = q.Where("listings.bathroom_count >= ?", *where.MinBathrooms)
	}
	if where.MinInteriorAreaSqft != nil {
		q = q.Where("listings.interior_area_sqft >= ?", *where.MinInteriorAreaSqft)
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

func (r *Repository) ListListingsForMap(ctx context.Context, where Where) ([]MapPinRow, error) {
	latestPrice := r.db.Model(&PriceHistory{}).
		Select("DISTINCT ON (board, mls) board, mls, price").
		Order("board, mls, observed_at DESC")

	q := r.db.WithContext(ctx).Model(&Listing{}).
		Select("listings.board, listings.mls, listings.latitude, listings.longitude, listings.address, listings.slug, listings.photo_url, listings.building_type, listings.bedroom_count, listings.bathroom_count, listings.interior_area_sqft, listings.commute_seconds_downtown, listings.first_seen_at, listings.is_available, lp.price AS current_price, (fav.user_id IS NOT NULL) AS is_favorite").
		Joins("LEFT JOIN (?) AS lp ON lp.board = listings.board AND lp.mls = listings.mls", latestPrice).
		Joins("LEFT JOIN favorites fav ON fav.board = listings.board AND fav.mls = listings.mls AND fav.user_id = ?", where.UserID).
		Where("listings.latitude <> 0 AND listings.longitude <> 0")

	if !where.ShowUnavailable {
		q = q.Where("listings.is_available = ?", true)
	}
	if where.FavoritesOnly {
		q = q.Where("fav.user_id IS NOT NULL")
	}
	if len(where.BuildingTypes) > 0 {
		q = q.Where("listings.building_type IN ?", where.BuildingTypes)
	}
	if where.MaxPrice != nil {
		q = q.Where("lp.price IS NOT NULL AND lp.price <= ?", *where.MaxPrice)
	}
	if where.MaxCommuteSec != nil {
		q = q.Where("listings.commute_seconds_downtown IS NOT NULL AND listings.commute_seconds_downtown <= ?", *where.MaxCommuteSec)
	}
	if where.NewSince != nil {
		q = q.Where("listings.first_seen_at >= ?", *where.NewSince)
	}
	if where.MinBedrooms != nil {
		q = q.Where("listings.bedroom_count >= ?", *where.MinBedrooms)
	}
	if where.MinBathrooms != nil {
		q = q.Where("listings.bathroom_count >= ?", *where.MinBathrooms)
	}
	if where.MinInteriorAreaSqft != nil {
		q = q.Where("listings.interior_area_sqft >= ?", *where.MinInteriorAreaSqft)
	}

	var rows []MapPinRow
	if err := q.Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) ListPendingCommute(ctx context.Context, refresh bool) ([]PendingCommute, error) {
	q := r.db.WithContext(ctx).Model(&Listing{}).
		Select("board, mls, latitude, longitude").
		Where("is_available = ?", true)
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

func (r *Repository) InvalidateIfStale(ctx context.Context, threshold time.Time) (int, error) {
	result := r.db.WithContext(ctx).Model(&Listing{}).Where("is_available = ? AND last_updated_at < ?", true, threshold).Update("is_available", false)
	return int(result.RowsAffected), result.Error
}
