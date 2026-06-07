package favorite

import (
	"time"

	"github.com/google/uuid"
)

type Favorite struct {
	UserID    uuid.UUID `gorm:"column:user_id;primaryKey"`
	Board     int       `gorm:"column:board;primaryKey"`
	MLS       int       `gorm:"column:mls;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Favorite) TableName() string { return "favorites" }

type Key struct {
	Board int
	MLS   int
}

type Page struct {
	Limit  int
	Offset int
}

type Sort struct {
	By  string
	Dir string
}

type Row struct {
	Board                  int       `gorm:"column:board"`
	MLS                    int       `gorm:"column:mls"`
	Latitude               float64   `gorm:"column:latitude"`
	Longitude              float64   `gorm:"column:longitude"`
	Address                string    `gorm:"column:address"`
	Slug                   string    `gorm:"column:slug"`
	BuildingType           int       `gorm:"column:building_type"`
	BedroomCount           int       `gorm:"column:bedroom_count"`
	BathroomCount          int       `gorm:"column:bathroom_count"`
	InteriorAreaSqft       float64   `gorm:"column:interior_area_sqft"`
	CommuteSecondsDowntown *int      `gorm:"column:commute_seconds_downtown"`
	IsAvailable            bool      `gorm:"column:is_available"`
	FirstSeenAt            time.Time `gorm:"column:first_seen_at"`
	FavoritedAt            time.Time `gorm:"column:favorited_at"`
	CurrentPrice           *float64  `gorm:"column:current_price"`
}
