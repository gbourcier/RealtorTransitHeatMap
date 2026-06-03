package favorite

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

var ErrListingNotFound = errors.New("favorite: listing not found")

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var sortColumns = map[string]string{
	"favorited_date":      "f.created_at",
	"listing_posted_date": "l.first_seen_at",
	"price":               "lp.price",
	"commute":             "l.commute_seconds_downtown",
}

func (r *Repository) Add(ctx context.Context, userID uuid.UUID, board, mls int) error {
	var count int64
	if err := r.db.WithContext(ctx).
		Table("listings").
		Where("board = ? AND mls = ?", board, mls).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrListingNotFound
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&Favorite{UserID: userID, Board: board, MLS: mls}).Error
}

func (r *Repository) Remove(ctx context.Context, userID uuid.UUID, board, mls int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND board = ? AND mls = ?", userID, board, mls).
		Delete(&Favorite{}).Error
}

func (r *Repository) RemoveBatch(ctx context.Context, userID uuid.UUID, keys []Key) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	tuples := make([][]any, len(keys))
	for i, k := range keys {
		tuples[i] = []any{k.Board, k.MLS}
	}
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND (board, mls) IN ?", userID, tuples).
		Delete(&Favorite{})
	return res.RowsAffected, res.Error
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, page Page, sort Sort) ([]Row, int64, error) {
	latestPrice := r.db.Table("listing_price_history").
		Select("DISTINCT ON (board, mls) board, mls, price").
		Order("board, mls, observed_at DESC")

	q := r.db.WithContext(ctx).
		Table("favorites AS f").
		Select("l.board, l.mls, l.latitude, l.longitude, l.address, l.slug, l.bedroom_count, l.bathroom_count, l.interior_area_sqft, l.commute_seconds_downtown, l.is_available, l.first_seen_at, f.created_at AS favorited_at, lp.price AS current_price").
		Joins("JOIN listings l ON l.board = f.board AND l.mls = f.mls").
		Joins("LEFT JOIN (?) AS lp ON lp.board = l.board AND lp.mls = l.mls", latestPrice).
		Where("f.user_id = ?", userID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	col := sortColumns[sort.By]
	if col == "" {
		col = "f.created_at"
	}

	var rows []Row
	err := q.Order(col + " " + sort.Dir + " NULLS LAST").
		Limit(page.Limit).
		Offset(page.Offset).
		Scan(&rows).Error

	return rows, total, err
}
