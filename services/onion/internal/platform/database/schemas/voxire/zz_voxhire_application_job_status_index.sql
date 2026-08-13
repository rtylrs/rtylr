ALTER TABLE voxire_career_applications
    ADD INDEX idx_career_apps_job_status (job_id, status);
