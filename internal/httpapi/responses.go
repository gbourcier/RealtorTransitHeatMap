package httpapi

type FetchPricesResponse struct {
	Listings []ListingDTO `json:"listings"`
}

type ListingDTO struct {
	ID      string `json:"id"`
	Price   uint64 `json:"price"`
	Address string `json:"address"`
}
