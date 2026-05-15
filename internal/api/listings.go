package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/go-chi/chi/v5"
)

type ListingService interface {
	ListListings(ctx context.Context, where listing.Where, page listing.Page, sort listing.Sort) ([]listing.ListingRow, int64, error)
	GetListing(ctx context.Context, board, mls int) (*listing.Listing, error)
}

type listingHandlers struct {
	svc ListingService
}

var validSortBy = map[string]bool{
	"price":                    true,
	"first_seen_at":            true,
	"commute_seconds_downtown": true,
}

func (h *listingHandlers) list(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePage(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "first_seen_at"
	}

	if sortBy == "commute_time" {
		sortBy = "commute_seconds_downtown"
	}

	if !validSortBy[sortBy] {
		writeError(w, http.StatusBadRequest, "invalid 'sortBy' param")
		return
	}

	sortDir := r.URL.Query().Get("sortDir")
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	rows, total, err := h.svc.ListListings(
		r.Context(),
		listing.Where{},
		listing.Page{Limit: limit, Offset: offset},
		listing.Sort{By: sortBy, Dir: sortDir},
	)
	if err != nil {
		slog.Error("listListings failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]ListingResponse, len(rows))
	for i := range rows {
		out[i] = listingFromRow(&rows[i])
	}
	writeJSON(w, http.StatusOK, PaginatedResponse[ListingResponse]{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *listingHandlers) get(w http.ResponseWriter, r *http.Request) {
	board, err := strconv.Atoi(chi.URLParam(r, "board"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'board' param")
		return
	}
	mls, err := strconv.Atoi(chi.URLParam(r, "mls"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'mls' param")
		return
	}

	l, err := h.svc.GetListing(r.Context(), board, mls)
	if err != nil {
		if errors.Is(err, listing.ErrNotFound) {
			writeError(w, http.StatusNotFound, "listing not found")
			return
		}
		slog.Error("getListing failed", "err", err, "board", board, "mls", mls)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, listingDetailFromModel(l))
}
