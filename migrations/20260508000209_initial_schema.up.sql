CREATE TABLE listings (
    board                     INTEGER             NOT NULL,
    mls                       INTEGER             NOT NULL,
    latitude                  DOUBLE PRECISION NOT NULL,
    longitude                 DOUBLE PRECISION NOT NULL,
    address                   TEXT,
    status                    TEXT             NOT NULL,
    commute_seconds_downtown  INTEGER,
    commute_computed_at       TIMESTAMPTZ,
    first_seen_at             TIMESTAMPTZ      NOT NULL DEFAULT now(),
    last_updated_at           TIMESTAMPTZ      NOT NULL DEFAULT now(),
    PRIMARY KEY (board, mls)
);

CREATE TABLE listing_price_history (
    board       INTEGER          NOT NULL,
    mls         INTEGER          NOT NULL,
    observed_at TIMESTAMPTZ   NOT NULL,
    price       NUMERIC(12,2) NOT NULL,
    PRIMARY KEY (board, mls, observed_at),
    FOREIGN KEY (board, mls) REFERENCES listings (board, mls) ON DELETE CASCADE
);

SELECT create_hypertable('listing_price_history', 'observed_at');

CREATE OR REPLACE FUNCTION set_last_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER listings_set_last_updated_at
    BEFORE UPDATE ON listings
    FOR EACH ROW
EXECUTE FUNCTION set_last_updated_at();
