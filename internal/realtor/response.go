package realtor

type asyncPropertySearchResponse struct {
	ErrorCode errorCode       `json:"ErrorCode"`
	Paging    paging          `json:"Paging"`
	Results   []listingResult `json:"Results"`
}

type errorCode struct {
	Id          int    `json:"Id"`
	Description string `json:"Description"`
	Status      string `json:"Status"`
	ProductName string `json:"ProductName"`
	Version     string `json:"Version"`
}

type paging struct {
	RecordsPerPage int `json:"RecordsPerPage"`
	CurrentPage    int `json:"CurrentPage"`
	TotalRecords   int `json:"TotalRecords"`
	TotalPages     int `json:"TotalPages"`
}

type listingResult struct {
	ID                 string       `json:"Id"`
	MlsNumber          int          `json:"MlsNumber,string"`
	StatusId           string       `json:"StatusId"`
	Property           property     `json:"Property"`
	Building           building     `json:"Building"`
	RelativeDetailsURL string       `json:"RelativeDetailsURL"`
	Individual         []individual `json:"Individual"`
}

type property struct {
	PriceUnformattedValue string  `json:"PriceUnformattedValue"`
	Address               address `json:"Address"`
	Photo                 []photo `json:"Photo"`
}

type photo struct {
	HighResPath string `json:"HighResPath"`
	MedResPath  string `json:"MedResPath"`
	LowResPath  string `json:"LowResPath"`
}

type building struct {
	BathroomTotal string `json:"BathroomTotal"`
	Bedrooms      string `json:"Bedrooms"`
	SizeInterior  string `json:"SizeInterior"`
	Type          string `json:"Type"`
}

type address struct {
	AddressText string `json:"AddressText"`
	Latitude    string `json:"Latitude"`
	Longitude   string `json:"Longitude"`
}

type individual struct {
	Organization organization `json:"Organization"`
}

type organization struct {
	OrganizationId int `json:"OrganizationId"`
}
