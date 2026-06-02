package schedule

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	ErrNotFound       = errors.New("schedule: not found")
	ErrInvalidCron    = errors.New("schedule: invalid cron expression")
	ErrInvalidJobType = errors.New("schedule: invalid job type")

	ErrPolygonRequired     = errors.New("schedule: polygon_wkt is required for scrape_realtor")
	ErrInvalidBuildingType = errors.New("schedule: invalid building type id")
	ErrInvalidRange        = errors.New("schedule: invalid range, expected min-max")
	ErrInvalidPrice        = errors.New("schedule: price_min must be <= price_max")
)

const (
	JobTypeScrapeRealtor = "scrape_realtor"
	JobTypeRefreshGtfs   = "refresh_gtfs"
)

var ValidBuildingTypeIDs = map[int]bool{1: true, 2: true, 3: true, 16: true, 17: true, 19: true}

var rangeRe = regexp.MustCompile(`^\d+-\d+$`)

type ScrapeParams struct {
	BuildingTypeID *int    `gorm:"column:building_type_id"`
	BedRange       *string `gorm:"column:bed_range"`
	BathRange      *string `gorm:"column:bath_range"`
	PriceMin       *int    `gorm:"column:price_min"`
	PriceMax       *int    `gorm:"column:price_max"`
	PolygonWKT     *string `gorm:"column:polygon_wkt"`
}

type Schedule struct {
	ID           uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Name         string    `gorm:"column:name;not null;uniqueIndex"`
	CronExpr     string    `gorm:"column:cron_expr;not null"`
	JobType      string    `gorm:"column:job_type;not null;default:scrape_realtor"`
	Enabled      bool      `gorm:"column:enabled;not null"`
	ScrapeParams `gorm:"embedded"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
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

func ValidateScrapeParams(p ScrapeParams) error {
	if p.PolygonWKT == nil || *p.PolygonWKT == "" {
		return ErrPolygonRequired
	}
	if p.BuildingTypeID != nil && !ValidBuildingTypeIDs[*p.BuildingTypeID] {
		return ErrInvalidBuildingType
	}
	if err := validateRange(p.BedRange); err != nil {
		return err
	}
	if err := validateRange(p.BathRange); err != nil {
		return err
	}
	if p.PriceMin != nil && p.PriceMax != nil && *p.PriceMin > *p.PriceMax {
		return ErrInvalidPrice
	}
	return nil
}

func validateRange(s *string) error {
	if s == nil || *s == "" {
		return nil
	}
	if !rangeRe.MatchString(*s) {
		return ErrInvalidRange
	}
	return nil
}
