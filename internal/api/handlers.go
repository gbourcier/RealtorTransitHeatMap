package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/scraperun"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/worker"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ScrapeService is the API's view of the worker. Kept here so this package
// doesn't depend on worker internals beyond what's needed for these handlers.
type ScrapeService interface {
	StartScrape() (uuid.UUID, error)
	GetRun(ctx context.Context, id uuid.UUID) (*scraperun.ScrapeRun, error)
}

type handlers struct {
	scrapes ScrapeService
}

func (h *handlers) startScrape(w http.ResponseWriter, r *http.Request) {
	id, err := h.scrapes.StartScrape()
	if err != nil {
		if errors.Is(err, worker.ErrBusy) {
			writeError(w, http.StatusConflict, "a scrape is already in progress")
			return
		}
		slog.Error("startScrape failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, StartScrapeResponse{RunID: id.String()})
}

func (h *handlers) getScrape(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	run, err := h.scrapes.GetRun(r.Context(), id)
	if err != nil {
		if errors.Is(err, scraperun.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		slog.Error("getScrape failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, scrapeRunFromModel(run))
}
