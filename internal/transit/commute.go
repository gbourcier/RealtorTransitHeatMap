package transit

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/listing"
)

var ErrBusy = errors.New("commute computer: computation already in progress")
var ErrNoStops = errors.New("commute computer: no transit stops loaded; run gtfs-precompute first")

type ListingRepo interface {
	ListPendingCommute(ctx context.Context, refresh bool) ([]listing.PendingCommute, error)
	GetPendingCommute(ctx context.Context, board, mls int) (*listing.PendingCommute, error)
	GetListing(ctx context.Context, board, mls int) (*listing.Listing, error)
	UpdateCommute(ctx context.Context, board, mls int, seconds int, computedAt time.Time) error
}

type StopRepo interface {
	List(ctx context.Context) ([]Stop, error)
}

type Config struct {
	NearestStops int
	WalkSpeedMps float64
	WalkDetour   float64
}

type CommuteComputer struct {
	stops    StopRepo
	listings ListingRepo
	cfg      Config
	rootCtx  context.Context
	wg       sync.WaitGroup
	busy     atomic.Bool
}

func NewCommuteComputer(stops StopRepo, listings ListingRepo, cfg Config) *CommuteComputer {
	return &CommuteComputer{stops: stops, listings: listings, cfg: cfg}
}

func (w *CommuteComputer) Bind(rootCtx context.Context) { w.rootCtx = rootCtx }
func (w *CommuteComputer) Wait()                        { w.wg.Wait() }

func (w *CommuteComputer) StartComputation(refresh bool) error {
	if !w.busy.CompareAndSwap(false, true) {
		return ErrBusy
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer w.busy.Store(false)
		if err := w.runBatch(w.rootCtx, refresh); err != nil {
			slog.Error("transit batch failed", "err", err)
		}
	}()
	return nil
}

func (w *CommuteComputer) ComputeOne(ctx context.Context, board, mls int, refresh bool) error {
	l, err := w.listings.GetListing(ctx, board, mls)
	if err != nil {
		return err
	}
	if !refresh && l.CommuteSecondsDowntown != nil {
		return nil
	}

	stops, err := w.stops.List(ctx)
	if err != nil {
		return err
	}
	if len(stops) == 0 {
		return ErrNoStops
	}

	seconds, ok := w.computeForListing(listing.PendingCommute{
		Board:     l.Board,
		MLS:       l.MLS,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
	}, stops)
	if !ok {
		return errors.New("no reachable transit stop")
	}
	return w.listings.UpdateCommute(ctx, board, mls, seconds, time.Now())
}

func (w *CommuteComputer) runBatch(ctx context.Context, refresh bool) error {
	t0 := time.Now()
	stops, err := w.stops.List(ctx)
	if err != nil {
		return err
	}
	if len(stops) == 0 {
		return ErrNoStops
	}

	pending, err := w.listings.ListPendingCommute(ctx, refresh)
	if err != nil {
		return err
	}

	slog.Info("transit batch start", "pending", len(pending), "stops", len(stops), "refresh", refresh)

	processed, skipped := 0, 0
	now := time.Now()
	for _, p := range pending {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		seconds, ok := w.computeForListing(p, stops)
		if !ok {
			skipped++
			continue
		}
		if err := w.listings.UpdateCommute(ctx, p.Board, p.MLS, seconds, now); err != nil {
			slog.Warn("update commute failed", "board", p.Board, "mls", p.MLS, "err", err)
			skipped++
			continue
		}
		processed++
	}

	slog.Info("transit batch done",
		"processed", processed,
		"skipped", skipped,
		"elapsed", time.Since(t0),
	)
	return nil
}

func (w *CommuteComputer) computeForListing(p listing.PendingCommute, stops []Stop) (int, bool) {
	if p.Latitude == 0 && p.Longitude == 0 {
		return 0, false
	}

	type scored struct {
		stop *Stop
		dist float64
	}
	candidates := make([]scored, 0, len(stops))
	for i := range stops {
		s := &stops[i]
		if s.CommuteSecondsToMcGill == nil {
			continue
		}
		d := haversineMeters(p.Latitude, p.Longitude, s.Latitude, s.Longitude)
		candidates = append(candidates, scored{stop: s, dist: d})
	}
	if len(candidates) == 0 {
		return 0, false
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].dist < candidates[j].dist })

	k := w.cfg.NearestStops
	if k > len(candidates) {
		k = len(candidates)
	}

	best := math.MaxInt32
	for i := 0; i < k; i++ {
		c := candidates[i]
		walkSec := int(c.dist * w.cfg.WalkDetour / w.cfg.WalkSpeedMps)
		total := walkSec + *c.stop.CommuteSecondsToMcGill
		if total < best {
			best = total
		}
	}
	if best == math.MaxInt32 {
		return 0, false
	}
	return best, true
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000.0
	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	dLa := (lat2 - lat1) * math.Pi / 180
	dLo := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLa/2)*math.Sin(dLa/2) + math.Cos(la1)*math.Cos(la2)*math.Sin(dLo/2)*math.Sin(dLo/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}
