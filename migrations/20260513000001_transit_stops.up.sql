CREATE TABLE transit_stops (
    stop_id                   TEXT             PRIMARY KEY,
    agency                    TEXT             NOT NULL,
    name                      TEXT             NOT NULL,
    latitude                  DOUBLE PRECISION NOT NULL,
    longitude                 DOUBLE PRECISION NOT NULL,
    commute_seconds_to_mcgill INTEGER,
    snapshot_key              TEXT             NOT NULL,
    computed_at               TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX transit_stops_lat_lon_idx ON transit_stops (latitude, longitude);
CREATE INDEX transit_stops_agency_idx ON transit_stops (agency);
