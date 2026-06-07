ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_building_type_check;

UPDATE listings
SET building_type = CASE building_type
    WHEN 0 THEN 1
    WHEN 1 THEN 2
    WHEN 2 THEN 3
    WHEN 3 THEN 19
    ELSE building_type
END
WHERE building_type IN (0, 1, 2, 3);

ALTER TABLE listings
    ADD CONSTRAINT listings_building_type_check
    CHECK (building_type IN (1, 2, 3, 17, 19));
