CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE scrape_runs (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    source        TEXT         NOT NULL,
    status        TEXT         NOT NULL,
    started_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    completed_at  TIMESTAMPTZ,
    total_count   INTEGER,
    new_count     INTEGER,
    error_kind    TEXT,
    error_message TEXT
);

CREATE INDEX scrape_runs_started_at_idx ON scrape_runs (started_at DESC);
