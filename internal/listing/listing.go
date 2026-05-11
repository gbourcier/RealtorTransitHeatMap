package listing

import "time"

type Listing struct {
	Board          int            `gorm:"column:board;primaryKey"`
	MLS            int            `gorm:"column:mls;primaryKey"`
	Latitude       float64
	Longitude      float64
	Address        string
	Status         string
	Slug           string         `gorm:"column:slug"`
	FirstSeenAt    time.Time      `gorm:"column:first_seen_at"`
	PriceHistories []PriceHistory `gorm:"foreignKey:Board,MLS;references:Board,MLS"`
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
}

type ListingRow struct {
	Listing
	CurrentPrice *float64
}
