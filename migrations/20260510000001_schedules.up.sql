CREATE TABLE schedules (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT         NOT NULL UNIQUE,
    cron_expr  TEXT         NOT NULL,
    source     TEXT         NOT NULL,
    enabled    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX schedules_enabled_idx ON schedules (enabled) WHERE enabled = TRUE;
