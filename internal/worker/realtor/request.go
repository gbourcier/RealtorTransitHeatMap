package realtor

import (
	"net/url"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

func searchValues(cfg *config.Config, c listing.FetchCriteria, page int) url.Values {
	v := url.Values{}
	v.Set("ZoomLevel", "11")
	v.Set("Sort", "6-D")
	v.Set("PropertyTypeGroupID", "1")
	v.Set("TransactionTypeId", "2")
	v.Set("PropertySearchTypeId", "1")
	v.Set("BuildingTypeId", "1")
	v.Set("Currency", "CAD")
	v.Set("IncludeHiddenListings", "false")
	v.Set("RecordsPerPage", "200")
	v.Set("ApplicationId", "1")
	v.Set("CultureId", "1")
	v.Set("Version", "7.0")
	v.Set("CurrentPage", strconv.Itoa(page))
	v.Set("PolygonWKT", cfg.PolygonWKT)

	setIfNonEmpty(v, "LatitudeMax", cfg.LatitudeMax)
	setIfNonEmpty(v, "LatitudeMin", cfg.LatitudeMin)
	setIfNonEmpty(v, "LongitudeMax", cfg.LongitudeMax)
	setIfNonEmpty(v, "LongitudeMin", cfg.LongitudeMin)

	priceMin := cfg.PriceMin
	if c.MinPrice != nil {
		priceMin = strconv.FormatUint(*c.MinPrice, 10)
	}
	setIfNonEmpty(v, "PriceMin", priceMin)

	priceMax := cfg.PriceMax
	if c.MaxPrice != nil {
		priceMax = strconv.FormatUint(*c.MaxPrice, 10)
	}
	setIfNonEmpty(v, "PriceMax", priceMax)

	setIfNonEmpty(v, "BedRange", cfg.BedRange)
	setIfNonEmpty(v, "BathRange", cfg.BathRange)

	return v
}

func setIfNonEmpty(v url.Values, key, val string) {
	if val != "" {
		v.Set(key, val)
	}
}
