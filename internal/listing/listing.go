package listing

import "time"

type Listing struct {
	Board          int `gorm:"column:board;primaryKey"`
	MLS            int `gorm:"column:mls;primaryKey"`
	Latitude       float64
	Longitude      float64
	Address        string
	Status         string
	PriceHistories []PriceHistory `gorm:"foreignKey:Board,MLS;references:Board,MLS"`
	Price          float64        `gorm:"-"` // transport field: carries fetched price for repository insertion
}

func (Listing) TableName() string { return "listings" }

type PriceHistory struct {
	Board      int       `gorm:"column:board;primaryKey"`
	MLS        int       `gorm:"column:mls;primaryKey"`
	ObservedAt time.Time `gorm:"column:observed_at;primaryKey"`
	Price      float64   `gorm:"column:price;type:numeric(12,2)"`
}

func (PriceHistory) TableName() string { return "listing_price_history" }
