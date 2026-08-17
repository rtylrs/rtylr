-- name: ListBenefitPlans :many
SELECT id, owner_id, business_id, name, type, provider, cost_employer, cost_employee, is_active, created_at
FROM hr_benefit_plans
WHERE business_id = ?
ORDER BY name ASC
LIMIT 100;

-- name: CreateBenefitPlan :execresult
INSERT INTO hr_benefit_plans (owner_id, business_id, name, type, provider, cost_employer, cost_employee, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetBenefitPlanIDForBusiness :one
SELECT id
FROM hr_benefit_plans
WHERE business_id = ? AND id = ?;

-- name: CreateBenefitEnrollment :execresult
INSERT INTO hr_benefit_enrollments (owner_id, business_id, employee_id, plan_id, status)
VALUES (?, ?, ?, ?, 'active');

-- name: GetBenefitEnrollmentByID :one
SELECT id, owner_id, business_id, employee_id, plan_id, status, enrolled_at, cancelled_at
FROM hr_benefit_enrollments
WHERE business_id = ? AND id = ?;

-- name: ListBenefitEnrollmentsByEmployee :many
SELECT id, owner_id, business_id, employee_id, plan_id, status, enrolled_at, cancelled_at
FROM hr_benefit_enrollments
WHERE business_id = ? AND employee_id = ?
ORDER BY enrolled_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountBenefitEnrollmentsByEmployee :one
SELECT COUNT(*)
FROM hr_benefit_enrollments
WHERE business_id = ? AND employee_id = ?;
