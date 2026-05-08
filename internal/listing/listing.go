package listing

type Listing struct {
	ID      string
	Price   uint64
	Address string
}

type FetchCriteria struct {
	City     string
	MinPrice *uint64
	MaxPrice *uint64
}
