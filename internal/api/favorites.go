package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/auth"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/favorite"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxFavoriteBatch = 200

type FavoriteService interface {
	Add(ctx context.Context, userID uuid.UUID, board, mls int) error
	Remove(ctx context.Context, userID uuid.UUID, board, mls int) error
	RemoveBatch(ctx context.Context, userID uuid.UUID, keys []favorite.Key) (int64, error)
	List(ctx context.Context, userID uuid.UUID, page favorite.Page, sort favorite.Sort) ([]favorite.Row, int64, error)
}

type favoriteHandlers struct {
	svc FavoriteService
}

var validFavoriteSortBy = map[string]bool{
	"favorited_date":      true,
	"listing_posted_date": true,
	"price":               true,
	"commute":             true,
}

type favoriteKey struct {
	Board int `json:"board"`
	MLS   int `json:"mls"`
}

type favoriteBatchRequest struct {
	Items []favoriteKey `json:"items"`
}

func requireUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	u := auth.UserFromContext(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return uuid.Nil, false
	}
	return u.ID, true
}

func (h *favoriteHandlers) add(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req favoriteKey
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Board <= 0 || req.MLS <= 0 {
		writeError(w, http.StatusBadRequest, "board and mls are required")
		return
	}
	if err := h.svc.Add(r.Context(), userID, req.Board, req.MLS); err != nil {
		if errors.Is(err, favorite.ErrListingNotFound) {
			writeError(w, http.StatusNotFound, "listing not found")
			return
		}
		slog.Error("addFavorite failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (h *favoriteHandlers) removeOne(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
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
	if err := h.svc.Remove(r.Context(), userID, board, mls); err != nil {
		slog.Error("removeFavorite failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *favoriteHandlers) removeBatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req favoriteBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "items is required")
		return
	}
	if len(req.Items) > maxFavoriteBatch {
		writeError(w, http.StatusBadRequest, "too many items")
		return
	}
	keys := make([]favorite.Key, 0, len(req.Items))
	for _, it := range req.Items {
		if it.Board <= 0 || it.MLS <= 0 {
			writeError(w, http.StatusBadRequest, "each item requires board and mls")
			return
		}
		keys = append(keys, favorite.Key{Board: it.Board, MLS: it.MLS})
	}
	deleted, err := h.svc.RemoveBatch(r.Context(), userID, keys)
	if err != nil {
		slog.Error("removeBatchFavorites failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"deleted": deleted})
}

func (h *favoriteHandlers) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	limit, offset, err := parsePage(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "favorited_date"
	}
	if !validFavoriteSortBy[sortBy] {
		writeError(w, http.StatusBadRequest, "invalid 'sortBy' param")
		return
	}

	sortDir := r.URL.Query().Get("sortDir")
	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	rows, total, err := h.svc.List(
		r.Context(),
		userID,
		favorite.Page{Limit: limit, Offset: offset},
		favorite.Sort{By: sortBy, Dir: sortDir},
	)
	if err != nil {
		slog.Error("listFavorites failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]FavoriteResponse, len(rows))
	for i := range rows {
		out[i] = favoriteFromRow(&rows[i])
	}
	writeJSON(w, http.StatusOK, PaginatedResponse[FavoriteResponse]{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}
