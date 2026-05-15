package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/scrape"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scraperun"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ScrapeService interface {
	StartScrape() (uuid.UUID, error)
	GetRun(ctx context.Context, id uuid.UUID) (*scraperun.ScrapeRun, error)
	ListRuns(ctx context.Context, where scraperun.Where, page scraperun.Page) ([]scraperun.ScrapeRun, int64, error)
}

type handlers struct {
	scrapes ScrapeService
}

func (h *handlers) startScrape(w http.ResponseWriter, r *http.Request) {
	id, err := h.scrapes.StartScrape()
	if err != nil {
		if errors.Is(err, scrape.ErrBusy) {
			writeError(w, http.StatusConflict, "a scrape is already in progress")
			return
		}
		slog.Error("startScrape failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, StartScrapeResponse{RunID: id.String()})
}

func (h *handlers) listScrapes(w http.ResponseWriter, r *http.Request) {
	where := scraperun.Where{}
	if v := r.URL.Query().Get("status"); v != "" {
		switch v {
		case scraperun.StatusRunning, scraperun.StatusSuccess, scraperun.StatusError:
			where.Status = &v
		default:
			writeError(w, http.StatusBadRequest, "invalid 'status' query param")
			return
		}
	}
	if v := r.URL.Query().Get("scheduleId"); v != "" {
		sid, err := uuid.Parse(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid 'scheduleId' query param")
			return
		}
		where.ScheduleID = &sid
	}
	limit, offset, err := parsePage(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	runs, total, err := h.scrapes.ListRuns(r.Context(), where, scraperun.Page{Limit: limit, Offset: offset})
	if err != nil {
		slog.Error("listScrapes failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]ScrapeRunResponse, len(runs))
	for i := range runs {
		out[i] = scrapeRunFromModel(&runs[i])
	}
	writeJSON(w, http.StatusOK, PaginatedResponse[ScrapeRunResponse]{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
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
