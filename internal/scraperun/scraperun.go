package scraperun

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusError   = "error"
)

const (
	ErrorKindForbidden = "forbidden"
	ErrorKindNullCount = "null_count"
	ErrorKindUnknown   = "unknown"
)

// SourceRealtor identifies scrapes against realtor.ca. Add more constants here
// when additional sources are introduced.
const SourceRealtor = "realtor.ca"

type ScrapeRun struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Source       string     `gorm:"column:source"`
	Status       string     `gorm:"column:status"`
	StartedAt    time.Time  `gorm:"column:started_at"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
	TotalCount   *int       `gorm:"column:total_count"`
	NewCount     *int       `gorm:"column:new_count"`
	ErrorKind    *string    `gorm:"column:error_kind"`
	ErrorMessage *string    `gorm:"column:error_message"`
}

func (ScrapeRun) TableName() string { return "scrape_runs" }
