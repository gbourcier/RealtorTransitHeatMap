ALTER TABLE schedules
    DROP COLUMN IF EXISTS building_type_id,
    DROP COLUMN IF EXISTS bed_range,
    DROP COLUMN IF EXISTS bath_range,
    DROP COLUMN IF EXISTS price_min,
    DROP COLUMN IF EXISTS price_max,
    DROP COLUMN IF EXISTS polygon_wkt;
