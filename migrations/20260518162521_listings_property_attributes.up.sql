ALTER TABLE listings ADD COLUMN bedroom_count       INTEGER          NOT NULL DEFAULT 0;
ALTER TABLE listings ADD COLUMN bathroom_count      INTEGER          NOT NULL DEFAULT 0;
ALTER TABLE listings ADD COLUMN interior_area_sqft  DOUBLE PRECISION NOT NULL DEFAULT 0;

ALTER TABLE listings ALTER COLUMN bedroom_count      DROP DEFAULT;
ALTER TABLE listings ALTER COLUMN bathroom_count     DROP DEFAULT;
ALTER TABLE listings ALTER COLUMN interior_area_sqft DROP DEFAULT;
