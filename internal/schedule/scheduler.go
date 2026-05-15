package schedule

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type Dispatcher interface {
	Dispatch(s Schedule) (uuid.UUID, error)
	IsBusy(err error) bool
}

type Scheduler struct {
	repo       *Repository
	dispatcher Dispatcher

	mu      sync.Mutex
	cron    *cron.Cron
	entries map[uuid.UUID]cron.EntryID
}

func New(repo *Repository, dispatcher Dispatcher) *Scheduler {
	return &Scheduler{
		repo:       repo,
		dispatcher: dispatcher,
		cron:       cron.New(cron.WithLocation(time.UTC)),
		entries:    make(map[uuid.UUID]cron.EntryID),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.Reload(ctx); err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

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
		entryID, err := s.cron.AddFunc(sch.CronExpr, s.fire(sch))
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

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) fire(sch Schedule) func() {
	return func() {
		runID, err := s.dispatcher.Dispatch(sch)
		switch {
		case s.dispatcher.IsBusy(err):
			slog.Info("scheduler: worker busy, skipping tick",
				"schedule_id", sch.ID, "name", sch.Name, "job_type", sch.JobType)
		case err != nil:
			slog.Error("scheduler: dispatch failed",
				"err", err, "schedule_id", sch.ID, "name", sch.Name, "job_type", sch.JobType)
		default:
			slog.Info("scheduler: dispatched",
				"schedule_id", sch.ID, "name", sch.Name, "job_type", sch.JobType, "run_id", runID)
		}
	}
}
