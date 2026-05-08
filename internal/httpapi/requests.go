package httpapi

type FetchPricesRequest struct {
	City     string  `json:"city"`
	MinPrice *uint64 `json:"min_price,omitempty"`
	MaxPrice *uint64 `json:"max_price,omitempty"`
}
