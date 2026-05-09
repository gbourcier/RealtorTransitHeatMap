package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

type FetchPricesService interface {
	FetchPrices(ctx context.Context) ([]listing.Listing, error)
}

type handlers struct {
	svc FetchPricesService
}

// fetchPrices godoc
// @Summary      Fetch listings
// @Description  Fetch real estate listings using the operator-configured search parameters
// @Tags         listings
// @Produce      json
// @Success      200 {object} FetchPricesResponse
// @Failure      408 {string} string "request cancelled"
// @Failure      500 {string} string "internal server error"
// @Router       /fetchPrices [get]
func (h *handlers) fetchPrices(w http.ResponseWriter, r *http.Request) {
	listings, err := h.svc.FetchPrices(r.Context())
	if err != nil {
		slog.Error("fetchPrices failed", "err", err)
		if errors.Is(err, context.Canceled) {
			http.Error(w, "request cancelled", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := FetchPricesResponse{Listings: make([]ListingDTO, 0, len(listings))}
	for _, l := range listings {
		out.Listings = append(out.Listings, ListingDTO{
			MLS:       l.MLS,
			Price:     l.Price,
			Address:   l.Address,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
