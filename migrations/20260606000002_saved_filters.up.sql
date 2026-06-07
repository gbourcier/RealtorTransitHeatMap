CREATE TABLE saved_filters (
    id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                   TEXT        NOT NULL,
    max_price              NUMERIC     CHECK (max_price >= 0),
    max_commute_sec        INTEGER     CHECK (max_commute_sec >= 0),
    new_within_days        INTEGER     CHECK (new_within_days > 0),
    min_bedrooms           INTEGER     CHECK (min_bedrooms BETWEEN 0 AND 20),
    min_bathrooms          INTEGER     CHECK (min_bathrooms BETWEEN 0 AND 20),
    min_interior_area_sqft NUMERIC     CHECK (min_interior_area_sqft >= 0),
    favorites_only         BOOLEAN     NOT NULL DEFAULT FALSE,
    include_expired        BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX saved_filters_user_name ON saved_filters (user_id, lower(name));
CREATE INDEX        saved_filters_user_idx  ON saved_filters (user_id, created_at);
