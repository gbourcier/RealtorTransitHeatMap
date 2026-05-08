package worker

import "github.com/gbourcier/RealtorTransitHeatMap/internal/listing"

type job struct {
	criteria listing.FetchCriteria
	reply    chan result
}

type result struct {
	listings []listing.Listing
	err      error
}
