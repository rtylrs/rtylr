-- Self-service benefit enrollments: one row per employee per benefit plan.
--
-- Scoped by business_id alongside owner_id, matching hr_benefit_plans. The
-- unique key makes a repeat enrollment in the same plan a conflict rather than
-- a silent duplicate; the API surfaces that as 409. There is no cancel RPC in
-- the contract yet, so 'cancelled' is currently only reachable by an operator.

CREATE TABLE IF NOT EXISTS hr_benefit_enrollments (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    business_id BIGINT NOT NULL DEFAULT 0,
    employee_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    status ENUM('active','cancelled') NOT NULL DEFAULT 'active',
    enrolled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cancelled_at TIMESTAMP NULL,
    UNIQUE KEY uniq_hr_benefit_enrollment (business_id, employee_id, plan_id),
    INDEX idx_hr_benefit_enrollments_employee (business_id, employee_id, enrolled_at),
    INDEX idx_hr_benefit_enrollments_plan (business_id, plan_id)
);
