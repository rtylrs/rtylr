ALTER TABLE activity_log
 ADD COLUMN source_event_id bigint unsigned NULL,
 ADD UNIQUE KEY uk_activity_source_event (source_event_id);
