package schedule

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

type Patch struct {
	Name         *string
	CronExpr     *string
	JobType      *string
	Enabled      *bool
	ScrapeParams *ScrapeParams
}

type Where struct {
	Enabled *bool
	JobType *string
}

type Page struct {
	Limit  int
	Offset int
}

func (r *Repository) List(ctx context.Context, where Where, page Page) ([]Schedule, int64, error) {
	q := r.db.WithContext(ctx).Model(&Schedule{})
	if where.Enabled != nil {
		q = q.Where("enabled = ?", *where.Enabled)
	}
	if where.JobType != nil {
		q = q.Where("job_type = ?", *where.JobType)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q = q.Order("created_at ASC")
	if page.Limit > 0 {
		q = q.Limit(page.Limit).Offset(page.Offset)
	}
	var out []Schedule
	if err := q.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*Schedule, error) {
	var s Schedule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Create(ctx context.Context, s *Schedule) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, patch Patch) (*Schedule, error) {
	updates := map[string]any{}
	if patch.Name != nil {
		updates["name"] = *patch.Name
	}
	if patch.CronExpr != nil {
		updates["cron_expr"] = *patch.CronExpr
	}
	if patch.JobType != nil {
		updates["job_type"] = *patch.JobType
	}
	if patch.Enabled != nil {
		updates["enabled"] = *patch.Enabled
	}
	if patch.ScrapeParams != nil {
		sp := patch.ScrapeParams
		updates["building_type_id"] = sp.BuildingTypeID
		updates["bed_range"] = sp.BedRange
		updates["bath_range"] = sp.BathRange
		updates["price_min"] = sp.PriceMin
		updates["price_max"] = sp.PriceMax
		updates["polygon_wkt"] = sp.PolygonWKT
	}
	if len(updates) == 0 {
		return r.Get(ctx, id)
	}
	res := r.db.WithContext(ctx).Model(&Schedule{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.Get(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Schedule{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
