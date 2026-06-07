package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/savedfilter"
	"github.com/google/uuid"
)

type PreferenceService interface {
	SetDefaultFilter(ctx context.Context, userID uuid.UUID, filterID *uuid.UUID) error
}

type preferenceHandlers struct {
	prefs   PreferenceService
	filters SavedFilterService
}

type preferenceRequest struct {
	DefaultFilterID *string `json:"defaultFilterId"`
}

func (h *preferenceHandlers) patch(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}
	var req preferenceRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	var filterID *uuid.UUID
	if req.DefaultFilterID != nil {
		id, err := uuid.Parse(*req.DefaultFilterID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid 'defaultFilterId'")
			return
		}
		if _, err := h.filters.GetByID(r.Context(), userID, id); err != nil {
			if errors.Is(err, savedfilter.ErrNotFound) {
				writeError(w, http.StatusNotFound, "saved filter not found")
				return
			}
			slog.Error("setDefaultFilter lookup failed", "err", err)
			writeError(w, http.StatusInternalServerError, "failed to update preferences")
			return
		}
		filterID = &id
	}

	if err := h.prefs.SetDefaultFilter(r.Context(), userID, filterID); err != nil {
		slog.Error("setDefaultFilter failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}
	writeJSON(w, http.StatusOK, preferenceRequest{DefaultFilterID: req.DefaultFilterID})
}
