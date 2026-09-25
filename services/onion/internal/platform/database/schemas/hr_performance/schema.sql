CREATE TABLE IF NOT EXISTS hr_performance_config (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    rating_scale_type VARCHAR(50) NOT NULL DEFAULT 'numeric',
    rating_min INT NOT NULL DEFAULT 1,
    rating_max INT NOT NULL DEFAULT 5,
    rating_labels JSON DEFAULT (JSON_ARRAY()),
    peer_feedback_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_perf_config_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_performance_review_cycles (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    cycle_type VARCHAR(50) NOT NULL DEFAULT 'annual',
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_perf_cycle_name (owner_id, name),
    INDEX idx_hr_perf_cycles_owner (owner_id)
);

CREATE TABLE IF NOT EXISTS hr_performance_reviews (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    review_cycle_id BIGINT NOT NULL,
    reviewer_employee_id BIGINT NOT NULL,
    review_type VARCHAR(50) NOT NULL DEFAULT 'self',
    overall_rating VARCHAR(100),
    narrative TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    shared_with_employee_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_hr_perf_review (owner_id, review_cycle_id, employee_id, reviewer_employee_id, review_type),
    INDEX idx_hr_perf_reviews_owner (owner_id),
    INDEX idx_hr_perf_reviews_employee (owner_id, employee_id),
    INDEX idx_hr_perf_reviews_cycle (owner_id, review_cycle_id)
);

CREATE TABLE IF NOT EXISTS hr_performance_review_goals (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    review_cycle_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    weight DECIMAL(5,2),
    success_criteria TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_hr_perf_goals_owner (owner_id),
    INDEX idx_hr_perf_goals_employee_cycle (owner_id, employee_id, review_cycle_id)
);
