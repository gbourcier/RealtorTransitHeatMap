package httpapi

type FetchPricesResponse struct {
	Listings []ListingDTO `json:"listings"`
}

type ListingDTO struct {
	MLS       string  `json:"mls"`
	Price     float64 `json:"price"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
