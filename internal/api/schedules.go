package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/schedule"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ScheduleRepo is the API's view of the schedule store.
type ScheduleRepo interface {
	List(ctx context.Context, where schedule.Where) ([]schedule.Schedule, error)
	Get(ctx context.Context, id uuid.UUID) (*schedule.Schedule, error)
	Create(ctx context.Context, s *schedule.Schedule) error
	Update(ctx context.Context, id uuid.UUID, patch schedule.Patch) (*schedule.Schedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ScheduleReloader is the in-memory cron runtime; handlers call Reload after
// writes so the active schedule table stays in sync with the DB.
type ScheduleReloader interface {
	Reload(ctx context.Context) error
}

type scheduleHandlers struct {
	repo     ScheduleRepo
	reloader ScheduleReloader
}

func (h *scheduleHandlers) list(w http.ResponseWriter, r *http.Request) {
	where := schedule.Where{}
	if v := r.URL.Query().Get("enabled"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid 'enabled' query param")
			return
		}
		where.Enabled = &b
	}
	items, err := h.repo.List(r.Context(), where)
	if err != nil {
		slog.Error("schedules list failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]ScheduleResponse, len(items))
	for i := range items {
		out[i] = scheduleFromModel(&items[i])
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *scheduleHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}
	s, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, schedule.ErrNotFound) {
			writeError(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.Error("schedules get failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, scheduleFromModel(s))
}

func (h *scheduleHandlers) create(w http.ResponseWriter, r *http.Request) {
	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := schedule.ValidateCron(req.CronExpr); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s := &schedule.Schedule{
		Name:     req.Name,
		CronExpr: req.CronExpr,
		Enabled:  true,
	}
	if req.Enabled != nil {
		s.Enabled = *req.Enabled
	}

	if err := h.repo.Create(r.Context(), s); err != nil {
		slog.Error("schedules create failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.reloader.Reload(r.Context()); err != nil {
		slog.Error("schedules reload failed", "err", err)
	}
	writeJSON(w, http.StatusCreated, scheduleFromModel(s))
}

func (h *scheduleHandlers) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}
	var req UpdateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.CronExpr != nil {
		if err := schedule.ValidateCron(*req.CronExpr); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if req.Name != nil && *req.Name == "" {
		writeError(w, http.StatusBadRequest, "name cannot be empty")
		return
	}

	s, err := h.repo.Update(r.Context(), id, schedule.Patch{
		Name:     req.Name,
		CronExpr: req.CronExpr,
		Enabled:  req.Enabled,
	})
	if err != nil {
		if errors.Is(err, schedule.ErrNotFound) {
			writeError(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.Error("schedules update failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.reloader.Reload(r.Context()); err != nil {
		slog.Error("schedules reload failed", "err", err)
	}
	writeJSON(w, http.StatusOK, scheduleFromModel(s))
}

func (h *scheduleHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid schedule id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, schedule.ErrNotFound) {
			writeError(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.Error("schedules delete failed", "err", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := h.reloader.Reload(r.Context()); err != nil {
		slog.Error("schedules reload failed", "err", err)
	}
	w.WriteHeader(http.StatusNoContent)
}
