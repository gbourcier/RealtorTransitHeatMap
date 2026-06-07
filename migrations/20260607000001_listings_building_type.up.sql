ALTER TABLE listings
    ADD COLUMN building_type INTEGER NOT NULL DEFAULT 0
        CHECK (building_type IN (0, 1, 2, 3));

ALTER TABLE listings ALTER COLUMN building_type DROP DEFAULT;
