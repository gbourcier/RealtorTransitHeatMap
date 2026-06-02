ALTER TABLE schedules
    ADD COLUMN building_type_id INTEGER,
    ADD COLUMN bed_range        TEXT,
    ADD COLUMN bath_range       TEXT,
    ADD COLUMN price_min        INTEGER,
    ADD COLUMN price_max        INTEGER,
    ADD COLUMN polygon_wkt      TEXT;
