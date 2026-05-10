package scraperun

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a scrape run lookup finds no matching row.
var ErrNotFound = errors.New("scraperun: not found")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Start inserts a new scrape_runs row with status=running and returns it.
// The returned row's StartedAt is the canonical run start time. Pass a
// non-nil scheduleID when the run is triggered by a schedule; pass nil
// for manual triggers.
func (r *Repository) Start(ctx context.Context, source string, scheduleID *uuid.UUID) (*ScrapeRun, error) {
	run := &ScrapeRun{
		ID:         uuid.New(),
		Source:     source,
		Status:     StatusRunning,
		StartedAt:  time.Now(),
		ScheduleID: scheduleID,
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

// Get returns the scrape run with the given id, or ErrNotFound if no row matches.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*ScrapeRun, error) {
	var run ScrapeRun
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}
