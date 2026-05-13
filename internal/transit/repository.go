package transit

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Replace(ctx context.Context, stops []Stop) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM transit_stops").Error; err != nil {
			return err
		}
		if len(stops) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{UpdateAll: true}).
			CreateInBatches(stops, 500).Error
	})
}

func (r *Repository) List(ctx context.Context) ([]Stop, error) {
	var stops []Stop
	err := r.db.WithContext(ctx).
		Where("commute_seconds_to_mcgill IS NOT NULL").
		Find(&stops).Error
	return stops, err
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&Stop{}).Count(&n).Error
	return n, err
}
