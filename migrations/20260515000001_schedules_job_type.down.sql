DROP INDEX IF EXISTS schedules_job_type_idx;

ALTER TABLE schedules DROP COLUMN IF EXISTS job_type;
