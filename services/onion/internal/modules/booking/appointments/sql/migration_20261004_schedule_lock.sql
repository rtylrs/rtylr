CREATE TABLE IF NOT EXISTS booking_schedule_locks (
 business_id BIGINT NOT NULL,
 PRIMARY KEY (business_id)
) ENGINE=InnoDB;
