package savedfilter

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("savedfilter: not found")
	ErrNameTaken = errors.New("savedfilter: name already exists")
)

type SavedFilter struct {
	ID                  uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	UserID              uuid.UUID `gorm:"column:user_id"`
	Name                string    `gorm:"column:name"`
	MaxPrice            *float64  `gorm:"column:max_price"`
	BuildingTypes       string    `gorm:"column:building_types"`
	MaxCommuteSec       *int      `gorm:"column:max_commute_sec"`
	NewWithinDays       *int      `gorm:"column:new_within_days"`
	MinBedrooms         *int      `gorm:"column:min_bedrooms"`
	MinBathrooms        *int      `gorm:"column:min_bathrooms"`
	MinInteriorAreaSqft *float64  `gorm:"column:min_interior_area_sqft"`
	FavoritesOnly       bool      `gorm:"column:favorites_only"`
	IncludeExpired      bool      `gorm:"column:include_expired"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

func (SavedFilter) TableName() string { return "saved_filters" }
