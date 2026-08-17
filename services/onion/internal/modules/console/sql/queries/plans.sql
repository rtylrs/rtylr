-- name: ListPlans :many
SELECT id, name, slug, price_per_seat_cents, annual_price_per_seat_cents, trial_days,
       max_seats, max_locations, is_active, is_solo, is_featured, tier_order, tagline, created_at
FROM plans
ORDER BY tier_order ASC
LIMIT 100;

-- name: GetPlanByID :one
SELECT id, name, slug, price_per_seat_cents, annual_price_per_seat_cents, trial_days,
       max_seats, max_locations, is_active, is_solo, is_featured, tier_order, tagline, created_at
FROM plans
WHERE id = ?
ORDER BY id ASC
LIMIT 1;

-- name: GetPlanBySlug :one
SELECT id, name, slug, price_per_seat_cents, annual_price_per_seat_cents, trial_days,
       max_seats, max_locations, is_active, is_solo, is_featured, tier_order, tagline, created_at
FROM plans
WHERE slug = ?
ORDER BY slug ASC
LIMIT 1;

-- name: UpdatePlan :exec
UPDATE plans
SET name = ?, price_per_seat_cents = ?, annual_price_per_seat_cents = ?, is_active = ?
WHERE id = ?;
