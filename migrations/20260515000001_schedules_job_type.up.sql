ALTER TABLE schedules
    ADD COLUMN job_type TEXT NOT NULL DEFAULT 'scrape_realtor'
        CHECK (job_type IN ('scrape_realtor', 'refresh_gtfs'));

CREATE INDEX schedules_job_type_idx ON schedules (job_type);
