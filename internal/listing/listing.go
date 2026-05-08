package listing

type Listing struct {
	Board     string
	MLS       string
	Latitude  float64
	Longitude float64
	Address   string
	Status    string
	Price     float64
}

type FetchCriteria struct {
	City     string
	MinPrice *uint64
	MaxPrice *uint64
}
