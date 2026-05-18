ALTER TABLE listings ADD COLUMN status TEXT NOT NULL DEFAULT '0';

UPDATE listings SET status = CASE WHEN is_available THEN '1' ELSE '0' END;

ALTER TABLE listings ALTER COLUMN status DROP DEFAULT;

ALTER TABLE listings DROP COLUMN is_available;
