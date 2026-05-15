package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/transit"
	"github.com/go-chi/chi/v5"
)

type TransitService interface {
	StartComputation(refresh bool) error
	ComputeOne(ctx context.Context, board, mls int, refresh bool) error
}

type transitHandlers struct {
	svc TransitService
}

func (h *transitHandlers) compute(w http.ResponseWriter, r *http.Request) {
	refresh := r.URL.Query().Get("refresh") == "true"
	if err := h.svc.StartComputation(refresh); err != nil {
		if errors.Is(err, transit.ErrBusy) {
			writeError(w, http.StatusConflict, "transit computation already in progress")
			return
		}
		if errors.Is(err, transit.ErrNoStops) {
			writeError(w, http.StatusFailedDependency, "no transit stops loaded; run gtfs-precompute first")
			return
		}
		slog.Error("transit compute failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"started": true, "refresh": refresh})
}

func (h *transitHandlers) computeOne(w http.ResponseWriter, r *http.Request) {
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
	refresh := r.URL.Query().Get("refresh") == "true"

	if err := h.svc.ComputeOne(r.Context(), board, mls, refresh); err != nil {
		if errors.Is(err, listing.ErrNotFound) {
			writeError(w, http.StatusNotFound, "listing not found")
			return
		}
		if errors.Is(err, transit.ErrNoStops) {
			writeError(w, http.StatusFailedDependency, "no transit stops loaded; run gtfs-precompute first")
			return
		}
		slog.Error("transit computeOne failed", "err", err, "board", board, "mls", mls)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"computed": true})
}
