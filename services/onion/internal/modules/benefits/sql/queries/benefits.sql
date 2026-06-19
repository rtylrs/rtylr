-- name: ListBenefitPlans :many
SELECT id, owner_id, business_id, name, type, provider, cost_employer, cost_employee, is_active, created_at
FROM hr_benefit_plans
WHERE business_id = ?
ORDER BY name ASC;

-- name: CreateBenefitPlan :execresult
INSERT INTO hr_benefit_plans (owner_id, business_id, name, type, provider, cost_employer, cost_employee, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
