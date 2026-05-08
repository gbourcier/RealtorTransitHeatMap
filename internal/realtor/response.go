package realtor

// asyncPropertySearchResponse mirrors the JSON shape returned by realtor.ca's
// PropertySearch_Post endpoint. These types are package-private on purpose:
// they describe realtor.ca's wire format and must not leak to other packages.
type asyncPropertySearchResponse struct {
	Paging  paging          `json:"Paging"`
	Results []listingResult `json:"Results"`
}

type paging struct {
	RecordsPerPage int `json:"RecordsPerPage"`
	CurrentPage    int `json:"CurrentPage"`
	TotalRecords   int `json:"TotalRecords"`
	TotalPages     int `json:"TotalPages"`
}

type listingResult struct {
	ID                 string   `json:"Id"`
	MlsNumber          string   `json:"MlsNumber"`
	Property           property `json:"Property"`
	RelativeDetailsURL string   `json:"RelativeDetailsURL"`
}

type property struct {
	PriceUnformattedValue string  `json:"PriceUnformattedValue"`
	Address               address `json:"Address"`
}

type address struct {
	AddressText string `json:"AddressText"`
	Latitude    string `json:"Latitude"`
	Longitude   string `json:"Longitude"`
}
