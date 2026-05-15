package refresh

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusError   = "error"
)

var ErrNotFound = errors.New("gtfs refresh: run not found")

type Run struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	ScheduleID   *uuid.UUID `gorm:"column:schedule_id;type:uuid"`
	Status       string     `gorm:"column:status"`
	StartedAt    time.Time  `gorm:"column:started_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
	StopsWritten *int       `gorm:"column:stops_written"`
	ErrorMessage *string    `gorm:"column:error_message"`
}

func (Run) TableName() string { return "gtfs_refresh_runs" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Start(ctx context.Context, scheduleID *uuid.UUID) (*Run, error) {
	run := &Run{
		ID:         uuid.New(),
		ScheduleID: scheduleID,
		Status:     StatusRunning,
		StartedAt:  time.Now(),
	}
	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

func (r *Repository) FinishSuccess(ctx context.Context, id uuid.UUID, stopsWritten int) (time.Time, error) {
	completedAt := time.Now()
	err := r.db.WithContext(ctx).
		Model(&Run{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        StatusSuccess,
			"completed_at":  completedAt,
			"stops_written": stopsWritten,
		}).Error
	return completedAt, err
}

func (r *Repository) FinishError(ctx context.Context, id uuid.UUID, message string) (time.Time, error) {
	completedAt := time.Now()
	err := r.db.WithContext(ctx).
		Model(&Run{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        StatusError,
			"completed_at":  completedAt,
			"error_message": message,
		}).Error
	return completedAt, err
}

type Where struct {
	Status     *string
	ScheduleID *uuid.UUID
}

type Page struct {
	Limit  int
	Offset int
}

func (r *Repository) List(ctx context.Context, where Where, page Page) ([]Run, int64, error) {
	q := r.db.WithContext(ctx).Model(&Run{})
	if where.Status != nil {
		q = q.Where("status = ?", *where.Status)
	}
	if where.ScheduleID != nil {
		q = q.Where("schedule_id = ?", *where.ScheduleID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	q = q.Order("started_at DESC")
	if page.Limit > 0 {
		q = q.Limit(page.Limit).Offset(page.Offset)
	}
	var out []Run
	if err := q.Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*Run, error) {
	var run Run
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &run, nil
}
