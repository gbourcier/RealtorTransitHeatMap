DROP INDEX IF EXISTS scrape_runs_schedule_id_idx;
ALTER TABLE scrape_runs DROP COLUMN IF EXISTS schedule_id;
