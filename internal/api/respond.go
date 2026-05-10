package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/schedule"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/scraperun"
)

type StartScrapeResponse struct {
	RunID string `json:"runId"`
}

type ScrapeRunResponse struct {
	ID           string  `json:"id"`
	Source       string  `json:"source"`
	Status       string  `json:"status"`
	StartedAt    int64   `json:"startedAt"`
	CompletedAt  *int64  `json:"completedAt,omitempty"`
	TotalCount   *int    `json:"totalCount,omitempty"`
	NewCount     *int    `json:"newCount,omitempty"`
	ErrorKind    string  `json:"errorKind,omitempty"`
	ErrorMessage string  `json:"errorMessage,omitempty"`
	ScheduleID   *string `json:"scheduleId,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// PaginatedResponse is the envelope returned by list endpoints. Total is the
// total matching rows (ignoring limit/offset) so the client can render page
// counts. Items is typed by the specific endpoint.
type PaginatedResponse[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

type ScheduleResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CronExpr  string `json:"cronExpr"`
	Enabled   bool   `json:"enabled"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type CreateScheduleRequest struct {
	Name     string `json:"name"`
	CronExpr string `json:"cronExpr"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

type UpdateScheduleRequest struct {
	Name     *string `json:"name,omitempty"`
	CronExpr *string `json:"cronExpr,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

func scheduleFromModel(s *schedule.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID:        s.ID.String(),
		Name:      s.Name,
		CronExpr:  s.CronExpr,
		Enabled:   s.Enabled,
		CreatedAt: s.CreatedAt.Unix(),
		UpdatedAt: s.UpdatedAt.Unix(),
	}
}

func scrapeRunFromModel(r *scraperun.ScrapeRun) ScrapeRunResponse {
	out := ScrapeRunResponse{
		ID:         r.ID.String(),
		Source:     r.Source,
		Status:     r.Status,
		StartedAt:  r.StartedAt.Unix(),
		TotalCount: r.TotalCount,
		NewCount:   r.NewCount,
	}
	if r.CompletedAt != nil {
		t := r.CompletedAt.Unix()
		out.CompletedAt = &t
	}
	if r.ErrorKind != nil {
		out.ErrorKind = *r.ErrorKind
	}
	if r.ErrorMessage != nil {
		out.ErrorMessage = *r.ErrorMessage
	}
	if r.ScheduleID != nil {
		s := r.ScheduleID.String()
		out.ScheduleID = &s
	}
	return out
}
