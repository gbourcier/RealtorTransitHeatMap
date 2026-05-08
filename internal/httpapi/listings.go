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
	FetchPrices(ctx context.Context, c listing.FetchCriteria) ([]listing.Listing, error)
}

type handlers struct {
	svc FetchPricesService
}

// fetchPrices godoc
// @Summary      Fetch listings
// @Description  Fetch real estate listings filtered by optional price range
// @Tags         listings
// @Accept       json
// @Produce      json
// @Param        request body FetchPricesRequest true "Search criteria"
// @Success      200 {object} FetchPricesResponse
// @Failure      400 {string} string "invalid json"
// @Failure      408 {string} string "request cancelled"
// @Failure      500 {string} string "internal server error"
// @Router       /fetchPrices [post]
func (h *handlers) fetchPrices(w http.ResponseWriter, r *http.Request) {
	var req FetchPricesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	listings, err := h.svc.FetchPrices(r.Context(), listing.FetchCriteria{
		City:     req.City,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
	})
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
