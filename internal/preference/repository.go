package preference

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DefaultFilterID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	var pref Preference
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&pref).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return pref.DefaultFilterID, nil
}

func (r *Repository) SetDefaultFilter(ctx context.Context, userID uuid.UUID, filterID *uuid.UUID) error {
	pref := Preference{UserID: userID, DefaultFilterID: filterID}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"default_filter_id": filterID,
				"updated_at":        time.Now(),
			}),
		}).
		Create(&pref).Error
}
