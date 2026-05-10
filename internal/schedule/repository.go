package schedule

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Patch is the set of mutable fields for an update. Nil pointer = leave
// unchanged. Name and CronExpr are validated by the caller (handler) before
// reaching the repository.
type Patch struct {
	Name     *string
	CronExpr *string
	Source   *string
	Enabled  *bool
}

// Where is the optional filter passed to List. Nil pointer = no filter on
// that field.
type Where struct {
	Enabled *bool
}

func (r *Repository) List(ctx context.Context, where Where) ([]Schedule, error) {
	q := r.db.WithContext(ctx).Order("created_at ASC")
	if where.Enabled != nil {
		q = q.Where("enabled = ?", *where.Enabled)
	}
	var out []Schedule
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
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
	if patch.Source != nil {
		updates["source"] = *patch.Source
	}
	if patch.Enabled != nil {
		updates["enabled"] = *patch.Enabled
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
