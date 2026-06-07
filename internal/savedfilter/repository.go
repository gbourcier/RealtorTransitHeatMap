package savedfilter

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]SavedFilter, error) {
	var rows []SavedFilter
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at").
		Find(&rows).Error
	return rows, err
}

func (r *Repository) Create(ctx context.Context, sf *SavedFilter) error {
	if sf.ID == uuid.Nil {
		sf.ID = uuid.New()
	}
	err := r.db.WithContext(ctx).Create(sf).Error
	if isUniqueViolation(err) {
		return ErrNameTaken
	}
	return err
}

func (r *Repository) Update(ctx context.Context, userID, id uuid.UUID, fields SavedFilter) error {
	res := r.db.WithContext(ctx).
		Model(&SavedFilter{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{
			"name":                   fields.Name,
			"max_price":              fields.MaxPrice,
			"max_commute_sec":        fields.MaxCommuteSec,
			"new_within_days":        fields.NewWithinDays,
			"min_bedrooms":           fields.MinBedrooms,
			"min_bathrooms":          fields.MinBathrooms,
			"min_interior_area_sqft": fields.MinInteriorAreaSqft,
			"favorites_only":         fields.FavoritesOnly,
			"include_expired":        fields.IncludeExpired,
			"updated_at":             time.Now(),
		})
	if isUniqueViolation(res.Error) {
		return ErrNameTaken
	}
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&SavedFilter{}).Error
}

func (r *Repository) GetByID(ctx context.Context, userID, id uuid.UUID) (*SavedFilter, error) {
	var sf SavedFilter
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&sf).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sf, nil
}
