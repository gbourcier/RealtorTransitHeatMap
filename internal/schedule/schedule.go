package schedule

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	ErrNotFound       = errors.New("schedule: not found")
	ErrInvalidCron    = errors.New("schedule: invalid cron expression")
	ErrInvalidJobType = errors.New("schedule: invalid job type")
)

const (
	JobTypeScrapeRealtor = "scrape_realtor"
	JobTypeRefreshGtfs   = "refresh_gtfs"
)

type Schedule struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null;uniqueIndex"`
	CronExpr  string    `gorm:"column:cron_expr;not null"`
	JobType   string    `gorm:"column:job_type;not null;default:scrape_realtor"`
	Enabled   bool      `gorm:"column:enabled;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Schedule) TableName() string { return "schedules" }

func ValidateCron(expr string) error {
	if _, err := cron.ParseStandard(expr); err != nil {
		return errors.Join(ErrInvalidCron, err)
	}
	return nil
}

func ValidateJobType(jt string) error {
	switch jt {
	case JobTypeScrapeRealtor, JobTypeRefreshGtfs:
		return nil
	default:
		return ErrInvalidJobType
	}
}
