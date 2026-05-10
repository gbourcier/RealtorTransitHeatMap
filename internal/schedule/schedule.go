package schedule

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	ErrNotFound    = errors.New("schedule: not found")
	ErrInvalidCron = errors.New("schedule: invalid cron expression")
	ErrInvalidName = errors.New("schedule: name is required")
)

type Schedule struct {
	ID   uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Name string    `gorm:"column:name;not null;uniqueIndex"`

	// CronExpr is a 5-field cron expression (minute hour dom month dow)
	// interpreted in UTC. Per-schedule timezones are not supported; convert
	// local times to UTC before storing (e.g. 9am Eastern -> "0 13 * * *").
	CronExpr string `gorm:"column:cron_expr;not null"`

	// Source matches scrape_runs.source. Today only scraperun.SourceRealtor.
	Source  string `gorm:"column:source;not null"`
	Enabled bool   `gorm:"column:enabled;not null;default:true"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Schedule) TableName() string { return "schedules" }

// ValidateCron parses an expression to confirm it's well-formed. Used by the
// API layer before write so bad input never reaches the DB.
func ValidateCron(expr string) error {
	if _, err := cron.ParseStandard(expr); err != nil {
		return errors.Join(ErrInvalidCron, err)
	}
	return nil
}
