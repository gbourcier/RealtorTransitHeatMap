package realtor

import (
	"net/url"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

// searchValues builds the form-encoded body for realtor.ca's
// PropertySearch_Post endpoint from internal fetch criteria.
func searchValues(c listing.FetchCriteria) url.Values {
	v := url.Values{}
	v.Set("CultureId", "1")
	v.Set("ApplicationId", "1")
	v.Set("RecordsPerPage", "100")
	v.Set("CurrentPage", "1")
	if c.City != "" {
		v.Set("KeywordsString", c.City)
	}
	if c.MinPrice != nil {
		v.Set("PriceMin", strconv.FormatUint(*c.MinPrice, 10))
	}
	if c.MaxPrice != nil {
		v.Set("PriceMax", strconv.FormatUint(*c.MaxPrice, 10))
	}
	return v
}
