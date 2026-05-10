ALTER TABLE scrape_runs
    ADD COLUMN schedule_id UUID
    REFERENCES schedules(id) ON DELETE SET NULL;

CREATE INDEX scrape_runs_schedule_id_idx
    ON scrape_runs (schedule_id, started_at DESC)
    WHERE schedule_id IS NOT NULL;
