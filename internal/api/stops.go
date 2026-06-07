package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/transit"
)

type StopService interface {
	List(ctx context.Context) ([]transit.Stop, error)
}

type stopHandlers struct {
	svc StopService
}

type StopResponse struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	CommuteSec int     `json:"commuteSec"`
}

func (h *stopHandlers) list(w http.ResponseWriter, r *http.Request) {
	stops, err := h.svc.List(r.Context())
	if err != nil {
		slog.Error("listStops failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list stops")
		return
	}
	out := make([]StopResponse, 0, len(stops))
	for i := range stops {
		s := &stops[i]
		if s.CommuteSecondsToMcGill == nil {
			continue
		}
		out = append(out, StopResponse{
			Latitude:   s.Latitude,
			Longitude:  s.Longitude,
			CommuteSec: *s.CommuteSecondsToMcGill,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
