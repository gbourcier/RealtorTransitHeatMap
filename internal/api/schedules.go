package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/schedule"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ScheduleRepo interface {
	List(ctx context.Context, where schedule.Where, page schedule.Page) ([]schedule.Schedule, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*schedule.Schedule, error)
	Create(ctx context.Context, s *schedule.Schedule) error
	Update(ctx context.Context, id uuid.UUID, patch schedule.Patch) (*schedule.Schedule, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ScheduleReloader interface {
	Reload(ctx context.Context) error
}

type ScheduleDispatcher interface {
	Dispatch(s schedule.Schedule) (uuid.UUID, error)
	IsBusy(err error) bool
}

type scheduleHandlers struct {
	repo       ScheduleRepo
	reloader   ScheduleReloader
	dispatcher ScheduleDispatcher
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
	if v := r.URL.Query().Get("jobType"); v != "" {
		if err := schedule.ValidateJobType(v); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		where.JobType = &v
	}
	limit, offset, err := parsePage(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, total, err := h.repo.List(r.Context(), where, schedule.Page{Limit: limit, Offset: offset})
	if err != nil {
		slog.Error("schedules list failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list schedules")
		return
	}
	out := make([]ScheduleResponse, len(items))
	for i := range items {
		out[i] = scheduleFromModel(&items[i])
	}
	writeJSON(w, http.StatusOK, PaginatedResponse[ScheduleResponse]{
		Items:  out,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
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
		writeError(w, http.StatusInternalServerError, "failed to get schedule")
		return
	}
	writeJSON(w, http.StatusOK, scheduleFromModel(s))
}

func (h *scheduleHandlers) create(w http.ResponseWriter, r *http.Request) {
	var req CreateScheduleRequest
	if !decodeJSONBody(w, r, &req) {
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
	jobType := req.JobType
	if jobType == "" {
		jobType = schedule.JobTypeScrapeRealtor
	}
	if err := schedule.ValidateJobType(jobType); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	s := &schedule.Schedule{
		Name:     req.Name,
		CronExpr: req.CronExpr,
		JobType:  jobType,
		Enabled:  true,
	}
	if req.Enabled != nil {
		s.Enabled = *req.Enabled
	}
	if jobType == schedule.JobTypeScrapeRealtor {
		sp, _ := req.toScrapeParams()
		if err := schedule.ValidateScrapeParams(sp); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.ScrapeParams = sp
	}

	if err := h.repo.Create(r.Context(), s); err != nil {
		slog.Error("schedules create failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create schedule")
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
	if !decodeJSONBody(w, r, &req) {
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
	if req.JobType != nil {
		if err := schedule.ValidateJobType(*req.JobType); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	patch := schedule.Patch{
		Name:     req.Name,
		CronExpr: req.CronExpr,
		JobType:  req.JobType,
		Enabled:  req.Enabled,
	}
	if sp, ok := req.toScrapeParams(); ok {
		if err := schedule.ValidateScrapeParams(sp); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		patch.ScrapeParams = &sp
	}

	s, err := h.repo.Update(r.Context(), id, patch)
	if err != nil {
		if errors.Is(err, schedule.ErrNotFound) {
			writeError(w, http.StatusNotFound, "schedule not found")
			return
		}
		slog.Error("schedules update failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update schedule")
		return
	}
	if err := h.reloader.Reload(r.Context()); err != nil {
		slog.Error("schedules reload failed", "err", err)
	}
	writeJSON(w, http.StatusOK, scheduleFromModel(s))
}

func (h *scheduleHandlers) run(w http.ResponseWriter, r *http.Request) {
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
		slog.Error("schedules run get failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to run schedule")
		return
	}
	runID, err := h.dispatcher.Dispatch(*s)
	if err != nil {
		if h.dispatcher.IsBusy(err) {
			writeError(w, http.StatusConflict, "a job is already in progress")
			return
		}
		slog.Error("schedules run dispatch failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to run schedule")
		return
	}
	writeJSON(w, http.StatusAccepted, RunScheduleResponse{RunID: runID.String()})
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
		writeError(w, http.StatusInternalServerError, "failed to delete schedule")
		return
	}
	if err := h.reloader.Reload(r.Context()); err != nil {
		slog.Error("schedules reload failed", "err", err)
	}
	w.WriteHeader(http.StatusNoContent)
}
