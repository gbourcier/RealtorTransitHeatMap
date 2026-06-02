package realtor

import (
	"net/url"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
)

type SearchParams struct {
	BuildingTypeID string
	BedRange       string
	BathRange      string
	PriceMin       string
	PriceMax       string
	PolygonWKT     string
}

func searchValues(cfg config.RealtorConfig, params SearchParams, page int) url.Values {
	v := url.Values{}
	v.Set("ZoomLevel", "11")
	v.Set("Sort", "6-D")
	v.Set("PropertyTypeGroupID", "1")
	v.Set("TransactionTypeId", "2")
	v.Set("PropertySearchTypeId", "1")
	v.Set("Currency", "CAD")
	v.Set("IncludeHiddenListings", "false")
	v.Set("RecordsPerPage", "200")
	v.Set("ApplicationId", "1")
	v.Set("CultureId", "1")
	v.Set("Version", "7.0")
	v.Set("CurrentPage", strconv.Itoa(page))
	v.Set("PolygonWKT", params.PolygonWKT)

	setIfNonEmpty(v, "BuildingTypeId", params.BuildingTypeID)

	setIfNonEmpty(v, "LatitudeMax", cfg.LatitudeMax)
	setIfNonEmpty(v, "LatitudeMin", cfg.LatitudeMin)
	setIfNonEmpty(v, "LongitudeMax", cfg.LongitudeMax)
	setIfNonEmpty(v, "LongitudeMin", cfg.LongitudeMin)

	setIfNonEmpty(v, "PriceMin", params.PriceMin)
	setIfNonEmpty(v, "PriceMax", params.PriceMax)

	setIfNonEmpty(v, "BedRange", params.BedRange)
	setIfNonEmpty(v, "BathRange", params.BathRange)

	return v
}

func setIfNonEmpty(v url.Values, key, val string) {
	if val != "" {
		v.Set(key, val)
	}
}
