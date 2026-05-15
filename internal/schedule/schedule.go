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
)

type Schedule struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null;uniqueIndex"`
	CronExpr  string    `gorm:"column:cron_expr;not null"`
	Enabled   bool      `gorm:"column:enabled;not null;default:true"`
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
