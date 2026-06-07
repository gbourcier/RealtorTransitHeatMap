package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/savedfilter"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const maxSavedFilterName = 60

type SavedFilterService interface {
	List(ctx context.Context, userID uuid.UUID) ([]savedfilter.SavedFilter, error)
	Create(ctx context.Context, sf *savedfilter.SavedFilter) error
	Update(ctx context.Context, userID, id uuid.UUID, fields savedfilter.SavedFilter) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	GetByID(ctx context.Context, userID, id uuid.UUID) (*savedfilter.SavedFilter, error)
}

type savedFilterHandlers struct {
	svc SavedFilterService
}

type savedFilterRequest struct {
	Name                string   `json:"name"`
	MaxPrice            *float64 `json:"maxPrice"`
	MaxCommuteSec       *int     `json:"maxCommuteSec"`
	NewWithinDays       *int     `json:"newWithinDays"`
	MinBedrooms         *int     `json:"minBedrooms"`
	MinBathrooms        *int     `json:"minBathrooms"`
	MinInteriorAreaSqft *float64 `json:"minInteriorAreaSqft"`
	FavoritesOnly       bool     `json:"favoritesOnly"`
	IncludeExpired      bool     `json:"includeExpired"`
}

func (req *savedFilterRequest) validate() (string, string) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return "", "name is required"
	}
	if len(name) > maxSavedFilterName {
		return "", "name is too long"
	}
	if req.MaxPrice != nil && *req.MaxPrice < 0 {
		return "", "maxPrice must be >= 0"
	}
	if req.MaxCommuteSec != nil && *req.MaxCommuteSec < 0 {
		return "", "maxCommuteSec must be >= 0"
	}
	if req.NewWithinDays != nil && *req.NewWithinDays <= 0 {
		return "", "newWithinDays must be > 0"
	}
	if req.MinBedrooms != nil && (*req.MinBedrooms < 0 || *req.MinBedrooms > 20) {
		return "", "minBedrooms must be between 0 and 20"
	}
	if req.MinBathrooms != nil && (*req.MinBathrooms < 0 || *req.MinBathrooms > 20) {
		return "", "minBathrooms must be between 0 and 20"
	}
	if req.MinInteriorAreaSqft != nil && *req.MinInteriorAreaSqft < 0 {
		return "", "minInteriorAreaSqft must be >= 0"
	}
	return name, ""
}

func (req *savedFilterRequest) toModel(name string) savedfilter.SavedFilter {
	return savedfilter.SavedFilter{
		Name:                name,
		MaxPrice:            req.MaxPrice,
		MaxCommuteSec:       req.MaxCommuteSec,
		NewWithinDays:       req.NewWithinDays,
		MinBedrooms:         req.MinBedrooms,
		MinBathrooms:        req.MinBathrooms,
		MinInteriorAreaSqft: req.MinInteriorAreaSqft,
		FavoritesOnly:       req.FavoritesOnly,
		IncludeExpired:      req.IncludeExpired,
	}
}

func (h *savedFilterHandlers) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	rows, err := h.svc.List(r.Context(), userID)
	if err != nil {
		slog.Error("listSavedFilters failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]SavedFilterResponse, len(rows))
	for i := range rows {
		out[i] = savedFilterFromModel(&rows[i])
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *savedFilterHandlers) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req savedFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	name, msg := req.validate()
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	sf := req.toModel(name)
	sf.UserID = userID
	if err := h.svc.Create(r.Context(), &sf); err != nil {
		if errors.Is(err, savedfilter.ErrNameTaken) {
			writeError(w, http.StatusConflict, "a saved filter with that name already exists")
			return
		}
		slog.Error("createSavedFilter failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, savedFilterFromModel(&sf))
}

func (h *savedFilterHandlers) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'id' param")
		return
	}
	var req savedFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	name, msg := req.validate()
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	if err := h.svc.Update(r.Context(), userID, id, req.toModel(name)); err != nil {
		switch {
		case errors.Is(err, savedfilter.ErrNotFound):
			writeError(w, http.StatusNotFound, "saved filter not found")
		case errors.Is(err, savedfilter.ErrNameTaken):
			writeError(w, http.StatusConflict, "a saved filter with that name already exists")
		default:
			slog.Error("updateSavedFilter failed", "err", err)
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	updated, err := h.svc.GetByID(r.Context(), userID, id)
	if err != nil {
		slog.Error("updateSavedFilter reload failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, savedFilterFromModel(updated))
}

func (h *savedFilterHandlers) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'id' param")
		return
	}
	if err := h.svc.Delete(r.Context(), userID, id); err != nil {
		slog.Error("deleteSavedFilter failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
