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
			ID:      l.ID,
			Price:   l.Price,
			Address: l.Address,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
