CREATE TABLE gtfs_refresh_runs (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id   UUID         REFERENCES schedules(id) ON DELETE SET NULL,
    status        TEXT         NOT NULL,
    started_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ,
    stops_written INTEGER,
    error_message TEXT
);

CREATE INDEX gtfs_refresh_runs_started_at_idx ON gtfs_refresh_runs (started_at DESC);

CREATE INDEX gtfs_refresh_runs_schedule_id_idx
    ON gtfs_refresh_runs (schedule_id, started_at DESC)
    WHERE schedule_id IS NOT NULL;
