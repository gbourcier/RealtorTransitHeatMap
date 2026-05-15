package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs/refresh"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GtfsRefreshService interface {
	Start() (uuid.UUID, error)
	GetRun(ctx context.Context, id uuid.UUID) (*refresh.Run, error)
	ListRuns(ctx context.Context, where refresh.Where, page refresh.Page) ([]refresh.Run, int64, error)
}

type gtfsRefreshHandlers struct {
	svc GtfsRefreshService
}

func (h *gtfsRefreshHandlers) start(w http.ResponseWriter, r *http.Request) {
	id, err := h.svc.Start()
	if err != nil {
		if errors.Is(err, refresh.ErrBusy) {
			writeError(w, http.StatusConflict, "a gtfs refresh is already in progress")
			return
		}
		slog.Error("startGtfsRefresh failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, StartGtfsRefreshResponse{RunID: id.String()})
}

func (h *gtfsRefreshHandlers) list(w http.ResponseWriter, r *http.Request) {
	where := refresh.Where{}
	if v := r.URL.Query().Get("status"); v != "" {
		switch v {
		case refresh.StatusRunning, refresh.StatusSuccess, refresh.StatusError:
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
	runs, total, err := h.svc.ListRuns(r.Context(), where, refresh.Page{Limit: limit, Offset: offset})
	if err != nil {
		slog.Error("listGtfsRefreshRuns failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]GtfsRefreshRunResponse, len(runs))
	for i := range runs {
		out[i] = gtfsRefreshRunFromModel(&runs[i])
	}
	writeJSON(w, http.StatusOK, PaginatedResponse[GtfsRefreshRunResponse]{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *gtfsRefreshHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}
	run, err := h.svc.GetRun(r.Context(), id)
	if err != nil {
		if errors.Is(err, refresh.ErrNotFound) {
			writeError(w, http.StatusNotFound, "run not found")
			return
		}
		slog.Error("getGtfsRefreshRun failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, gtfsRefreshRunFromModel(run))
}
