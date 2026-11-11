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

-- name: ListPlanRegionPrices :many
SELECT plan_id, region, currency, price_per_seat_cents, annual_price_per_seat_cents,
       COALESCE(paddle_price_id, '') AS paddle_price_id, is_active
FROM plan_region_prices
ORDER BY region ASC, plan_id ASC;

-- name: UpsertPlanRegionPrice :exec
INSERT INTO plan_region_prices
  (plan_id, region, currency, price_per_seat_cents, annual_price_per_seat_cents, paddle_price_id, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  currency = VALUES(currency),
  price_per_seat_cents = VALUES(price_per_seat_cents),
  annual_price_per_seat_cents = VALUES(annual_price_per_seat_cents),
  paddle_price_id = VALUES(paddle_price_id),
  is_active = VALUES(is_active);

-- name: DeletePlanRegionPrice :execrows
DELETE FROM plan_region_prices WHERE plan_id = ? AND region = ?;
