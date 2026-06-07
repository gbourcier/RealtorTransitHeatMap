ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_building_type_check;

UPDATE listings
SET building_type = CASE building_type
    WHEN 1 THEN 0
    WHEN 2 THEN 1
    WHEN 3 THEN 2
    WHEN 19 THEN 3
    ELSE 0
END;

ALTER TABLE listings
    ADD CONSTRAINT listings_building_type_check
    CHECK (building_type IN (0, 1, 2, 3));
