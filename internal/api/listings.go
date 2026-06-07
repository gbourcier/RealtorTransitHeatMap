package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/auth"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/go-chi/chi/v5"
)

func parseListingWhere(r *http.Request) (listing.Where, error) {
	where := listing.Where{}
	if v := r.URL.Query().Get("maxPrice"); v != "" {
		n, perr := strconv.ParseFloat(v, 64)
		if perr != nil || n < 0 {
			return where, fmt.Errorf("invalid 'maxPrice' query param")
		}
		where.MaxPrice = &n
	}
	if v := r.URL.Query().Get("maxCommuteSec"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 0 {
			return where, fmt.Errorf("invalid 'maxCommuteSec' query param")
		}
		where.MaxCommuteSec = &n
	}
	if v := r.URL.Query().Get("newWithinDays"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 1 {
			return where, fmt.Errorf("invalid 'newWithinDays' query param")
		}
		since := time.Now().AddDate(0, 0, -n)
		where.NewSince = &since
	}
	if v := r.URL.Query().Get("minBedrooms"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 0 {
			return where, fmt.Errorf("invalid 'minBedrooms' query param")
		}
		where.MinBedrooms = &n
	}
	if v := r.URL.Query().Get("minBathrooms"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 0 {
			return where, fmt.Errorf("invalid 'minBathrooms' query param")
		}
		where.MinBathrooms = &n
	}
	if v := r.URL.Query().Get("minInteriorAreaSqft"); v != "" {
		n, perr := strconv.ParseFloat(v, 64)
		if perr != nil || n < 0 {
			return where, fmt.Errorf("invalid 'minInteriorAreaSqft' query param")
		}
		where.MinInteriorAreaSqft = &n
	}
	if r.URL.Query().Get("favoritesOnly") == "true" {
		where.FavoritesOnly = true
	}
	if r.URL.Query().Get("includeExpired") == "true" {
		where.ShowUnavailable = true
	}
	if u := auth.UserFromContext(r); u != nil {
		where.UserID = u.ID
	}
	return where, nil
}

type ListingService interface {
	ListListings(ctx context.Context, where listing.Where, page listing.Page, sort listing.Sort) ([]listing.ListingRow, int64, error)
	ListListingsForMap(ctx context.Context, where listing.Where) ([]listing.MapPinRow, error)
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

	where, werr := parseListingWhere(r)
	if werr != nil {
		writeError(w, http.StatusBadRequest, werr.Error())
		return
	}

	rows, total, err := h.svc.ListListings(
		r.Context(),
		where,
		listing.Page{Limit: limit, Offset: offset},
		listing.Sort{By: sortBy, Dir: sortDir},
	)
	if err != nil {
		slog.Error("listListings failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list listings")
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

func (h *listingHandlers) mapList(w http.ResponseWriter, r *http.Request) {
	where, werr := parseListingWhere(r)
	if werr != nil {
		writeError(w, http.StatusBadRequest, werr.Error())
		return
	}

	rows, err := h.svc.ListListingsForMap(r.Context(), where)
	if err != nil {
		slog.Error("listListingsForMap failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list map listings")
		return
	}

	out := make([]ListingMapPinResponse, len(rows))
	for i := range rows {
		out[i] = mapPinFromRow(&rows[i])
	}
	writeJSON(w, http.StatusOK, out)
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
		writeError(w, http.StatusInternalServerError, "failed to get listing")
		return
	}
	writeJSON(w, http.StatusOK, listingDetailFromModel(l))
}
