package schedule

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/worker"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// WorkerTrigger is the scheduler's view of the worker. Decoupled via interface
// so this package only depends on what it actually uses.
type WorkerTrigger interface {
	StartScrapeForSchedule(scheduleID uuid.UUID) (uuid.UUID, error)
}

type Scheduler struct {
	repo    *Repository
	trigger WorkerTrigger

	mu      sync.Mutex
	cron    *cron.Cron
	entries map[uuid.UUID]cron.EntryID
}

func New(repo *Repository, trigger WorkerTrigger) *Scheduler {
	return &Scheduler{
		repo:    repo,
		trigger: trigger,
		cron:    cron.New(cron.WithLocation(time.UTC)),
		entries: make(map[uuid.UUID]cron.EntryID),
	}
}

// Start loads enabled schedules from the DB, registers them with the cron
// runtime, and begins firing. ctx is used for the initial load only; the
// cron runtime runs until Stop is called.
func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.Reload(ctx); err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

// Reload re-reads enabled schedules and rebuilds the cron table to match.
// Call after any CRUD write to a schedule.
func (s *Scheduler) Reload(ctx context.Context) error {
	enabled := true
	schedules, _, err := s.repo.List(ctx, Where{Enabled: &enabled}, Page{})
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entryID := range s.entries {
		s.cron.Remove(entryID)
	}
	s.entries = make(map[uuid.UUID]cron.EntryID, len(schedules))

	for _, sch := range schedules {
		entryID, err := s.cron.AddFunc(sch.CronExpr, s.fire(sch.ID, sch.Name))
		if err != nil {
			slog.Error("scheduler: skipping invalid cron",
				"schedule_id", sch.ID, "name", sch.Name, "cron_expr", sch.CronExpr, "err", err)
			continue
		}
		s.entries[sch.ID] = entryID
	}
	slog.Info("scheduler: reloaded", "active", len(s.entries))
	return nil
}

// Stop halts the cron runtime. In-flight fires (already-running goroutines)
// are not cancelled here; the worker handles its own shutdown via Worker.Wait.
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) fire(scheduleID uuid.UUID, name string) func() {
	return func() {
		runID, err := s.trigger.StartScrapeForSchedule(scheduleID)
		switch {
		case errors.Is(err, worker.ErrBusy):
			slog.Info("scheduler: worker busy, skipping tick",
				"schedule_id", scheduleID, "name", name)
		case err != nil:
			slog.Error("scheduler: trigger failed",
				"err", err, "schedule_id", scheduleID, "name", name)
		default:
			slog.Info("scheduler: triggered scrape",
				"schedule_id", scheduleID, "name", name, "run_id", runID)
		}
	}
}
