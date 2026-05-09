package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/worker"
)

type ExecuteScrapeService interface {
	ExecuteScrape(ctx context.Context) (worker.ScrapeResult, error)
}

type handlers struct {
	svc ExecuteScrapeService
}

func (h *handlers) executeScrape(w http.ResponseWriter, r *http.Request) {
	res, err := h.svc.ExecuteScrape(r.Context())
	if err != nil {
		slog.Error("executeScrape failed", "err", err)
		if errors.Is(err, context.Canceled) {
			http.Error(w, "request cancelled", http.StatusRequestTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := ExecuteScrapeResponse{
		TotalCount:  res.TotalCount,
		NewCount:    res.NewCount,
		StartedAt:   res.StartedAt.Unix(),
		CompletedAt: res.CompletedAt.Unix(),
		ErrorKind:   res.ErrorKind,
		Error:       res.Error,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
