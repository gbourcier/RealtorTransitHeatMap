package scraperun

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Start inserts a new scrape_runs row with status=running and returns it.
// The returned row's StartedAt is the canonical run start time.
func (r *Repository) Start(ctx context.Context, source string) (*ScrapeRun, error) {
	run := &ScrapeRun{
		ID:        uuid.New(),
		Source:    source,
		Status:    StatusRunning,
		StartedAt: time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

// FinishSuccess marks the run as successful and writes the observed counts.
func (r *Repository) FinishSuccess(ctx context.Context, id uuid.UUID, totalCount, newCount int) (time.Time, error) {
	completedAt := time.Now()
	err := r.db.WithContext(ctx).
		Model(&ScrapeRun{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       StatusSuccess,
			"completed_at": completedAt,
			"total_count":  totalCount,
			"new_count":    newCount,
		}).Error
	return completedAt, err
}

// FinishError marks the run as errored. totalCount/newCount may be zero when
// the failure happened before any results were retrieved or persisted.
func (r *Repository) FinishError(ctx context.Context, id uuid.UUID, kind, message string, totalCount, newCount int) (time.Time, error) {
	completedAt := time.Now()
	err := r.db.WithContext(ctx).
		Model(&ScrapeRun{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        StatusError,
			"completed_at":  completedAt,
			"total_count":   totalCount,
			"new_count":     newCount,
			"error_kind":    kind,
			"error_message": message,
		}).Error
	return completedAt, err
}
