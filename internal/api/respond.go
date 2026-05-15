package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/gtfs/refresh"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
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
	JobType   string `json:"jobType"`
	Enabled   bool   `json:"enabled"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type CreateScheduleRequest struct {
	Name     string `json:"name"`
	CronExpr string `json:"cronExpr"`
	JobType  string `json:"jobType"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

type UpdateScheduleRequest struct {
	Name     *string `json:"name,omitempty"`
	CronExpr *string `json:"cronExpr,omitempty"`
	JobType  *string `json:"jobType,omitempty"`
	Enabled  *bool   `json:"enabled,omitempty"`
}

func scheduleFromModel(s *schedule.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID:        s.ID.String(),
		Name:      s.Name,
		CronExpr:  s.CronExpr,
		JobType:   s.JobType,
		Enabled:   s.Enabled,
		CreatedAt: s.CreatedAt.Unix(),
		UpdatedAt: s.UpdatedAt.Unix(),
	}
}

type StartGtfsRefreshResponse struct {
	RunID string `json:"runId"`
}

type GtfsRefreshRunResponse struct {
	ID           string  `json:"id"`
	Status       string  `json:"status"`
	StartedAt    int64   `json:"startedAt"`
	CompletedAt  *int64  `json:"completedAt,omitempty"`
	StopsWritten *int    `json:"stopsWritten,omitempty"`
	ErrorMessage string  `json:"errorMessage,omitempty"`
	ScheduleID   *string `json:"scheduleId,omitempty"`
}

func gtfsRefreshRunFromModel(r *refresh.Run) GtfsRefreshRunResponse {
	out := GtfsRefreshRunResponse{
		ID:           r.ID.String(),
		Status:       r.Status,
		StartedAt:    r.StartedAt.Unix(),
		StopsWritten: r.StopsWritten,
	}
	if r.CompletedAt != nil {
		t := r.CompletedAt.Unix()
		out.CompletedAt = &t
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

type ListingResponse struct {
	Board                  int      `json:"board"`
	MLS                    int      `json:"mls"`
	Address                string   `json:"address"`
	CurrentPrice           *float64 `json:"currentPrice"`
	CommuteSecondsDowntown *int     `json:"commuteSecondsDowntown,omitempty"`
	FirstSeenAt            int64    `json:"firstSeenAt"`
	Slug                   string   `json:"slug"`
}

type PriceHistoryResponse struct {
	ObservedAt int64   `json:"observedAt"`
	Price      float64 `json:"price"`
}

type ListingDetailResponse struct {
	ListingResponse
	Status         string                 `json:"status"`
	PriceHistories []PriceHistoryResponse `json:"priceHistories"`
}

func mapStatus(raw string) string {
	if raw == "1" {
		return "available"
	}
	return "unavailable"
}

func listingFromRow(row *listing.ListingRow) ListingResponse {
	return ListingResponse{
		Board:                  row.Board,
		MLS:                    row.MLS,
		Address:                row.Address,
		CurrentPrice:           row.CurrentPrice,
		CommuteSecondsDowntown: row.CommuteSecondsDowntown,
		FirstSeenAt:            row.FirstSeenAt.Unix(),
		Slug:                   "https://www.realtor.ca" + row.Slug,
	}
}

func listingDetailFromModel(l *listing.Listing) ListingDetailResponse {
	histories := make([]PriceHistoryResponse, len(l.PriceHistories))
	for i, ph := range l.PriceHistories {
		histories[i] = PriceHistoryResponse{
			ObservedAt: ph.ObservedAt.Unix(),
			Price:      ph.Price,
		}
	}

	var currentPrice *float64
	if len(l.PriceHistories) > 0 {
		p := l.PriceHistories[0].Price
		currentPrice = &p
	}
	return ListingDetailResponse{
		ListingResponse: ListingResponse{
			Board:                  l.Board,
			MLS:                    l.MLS,
			Address:                l.Address,
			CurrentPrice:           currentPrice,
			CommuteSecondsDowntown: l.CommuteSecondsDowntown,
			FirstSeenAt:            l.FirstSeenAt.Unix(),
			Slug:                   "https://www.realtor.ca" + l.Slug,
		},
		Status:         mapStatus(l.Status),
		PriceHistories: histories,
	}
}
