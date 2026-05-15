package listing

import "time"

type Listing struct {
	Board                  int            `gorm:"column:board;primaryKey"`
	MLS                    int            `gorm:"column:mls;primaryKey"`
	Latitude               float64
	Longitude              float64
	Address                string
	Status                 string
	Slug                   string         `gorm:"column:slug"`
	CommuteSecondsDowntown *int           `gorm:"column:commute_seconds_downtown"`
	CommuteComputedAt      *time.Time     `gorm:"column:commute_computed_at"`
	FirstSeenAt            time.Time      `gorm:"column:first_seen_at"`
	PriceHistories         []PriceHistory `gorm:"foreignKey:Board,MLS;references:Board,MLS"`
}

func (Listing) TableName() string { return "listings" }

type PriceHistory struct {
	Board      int       `gorm:"column:board;primaryKey"`
	MLS        int       `gorm:"column:mls;primaryKey"`
	ObservedAt time.Time `gorm:"column:observed_at;primaryKey"`
	Price      float64   `gorm:"column:price;type:numeric(12,2)"`
}

func (PriceHistory) TableName() string { return "listing_price_history" }

type Observation struct {
	Listing Listing
	Price   float64
}

type Page struct {
	Limit  int
	Offset int
}

type Sort struct {
	By  string
	Dir string
}

type Where struct {
	ShowUnavailable bool
	MaxPrice        *float64
	MaxCommuteSec   *int
	NewSince        *time.Time
}

type ListingRow struct {
	Listing
	CurrentPrice *float64
}

type PendingCommute struct {
	Board     int     `gorm:"column:board"`
	MLS       int     `gorm:"column:mls"`
	Latitude  float64 `gorm:"column:latitude"`
	Longitude float64 `gorm:"column:longitude"`
}
