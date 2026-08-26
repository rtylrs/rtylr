-- name: GetCustomerByID :one
SELECT id, owner_id, business_id, first_name, last_name, email, phone, date_of_birth, gender,
       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
	       loyalty_tier, tags, notes, consent, is_active, created_at, updated_at
FROM crm_customers WHERE id = ? AND business_id = ?;

-- name: CreateCustomer :execresult
INSERT INTO crm_customers
  (owner_id, business_id, first_name, last_name, email, phone, date_of_birth, gender,
	   address, city, country, tags, notes, consent, loyalty_tier, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'bronze', TRUE);

-- name: UpdateCustomer :execresult
UPDATE crm_customers
SET first_name = ?, last_name = ?, email = ?, phone = ?, date_of_birth = ?,
	    gender = ?, address = ?, city = ?, country = ?, tags = ?, notes = ?, consent = ?, is_active = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteCustomer :execresult
DELETE FROM crm_customers WHERE id = ? AND business_id = ?;

-- name: UpdateCustomerLoyalty :exec
UPDATE crm_customers SET loyalty_points = ?, loyalty_tier = ? WHERE id = ? AND business_id = ?;

-- name: RecordVisit :exec
UPDATE crm_customers
SET visit_count = visit_count + 1,
    last_visit_at = NOW(),
    total_spent = total_spent + ?
WHERE id = ? AND business_id = ?;

-- name: CountCustomers :one
SELECT COUNT(*) FROM crm_customers WHERE business_id = ?;

-- name: ListCustomers :many
SELECT id, owner_id, business_id, first_name, last_name, email, phone, date_of_birth, gender,
       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
	       loyalty_tier, tags, notes, consent, is_active, created_at, updated_at
FROM crm_customers WHERE business_id = ?
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: CountActiveCustomers :one
SELECT COUNT(*) FROM crm_customers WHERE business_id = ? AND is_active = TRUE;

-- name: GetBusinessTimezone :one
SELECT COALESCE(timezone, 'UTC') AS timezone
FROM businesses
WHERE id = ? AND is_active = TRUE;

-- name: CountNewCustomersThisMonth :one
SELECT COUNT(*) FROM crm_customers
WHERE business_id = ?
  AND created_at >= sqlc.arg(created_at_start)
  AND created_at < sqlc.arg(created_at_end);

-- name: CountActiveLoyaltyMembers :one
SELECT COUNT(*) FROM crm_customers WHERE business_id = ? AND loyalty_points > 0 AND is_active = TRUE;

-- name: ListTopCustomers :many
SELECT id, first_name, last_name, email, total_spent, loyalty_points, loyalty_tier
FROM crm_customers WHERE business_id = ? AND is_active = TRUE
ORDER BY total_spent DESC LIMIT 10;

-- name: ListActiveCustomers :many
SELECT id, owner_id, business_id, first_name, last_name, email, phone, date_of_birth, gender,
       address, city, country, loyalty_points, total_spent, visit_count, last_visit_at,
	       loyalty_tier, tags, notes, consent, is_active, created_at, updated_at
FROM crm_customers WHERE business_id = ? AND is_active = TRUE
ORDER BY created_at DESC LIMIT ? OFFSET ?;

-- name: UpdateCustomerMergeStats :exec
UPDATE crm_customers
SET loyalty_points = ?, total_spent = ?, visit_count = ?
WHERE id = ? AND business_id = ?;

-- name: MigrateLoyaltyTransactions :exec
UPDATE crm_loyalty_transactions SET customer_id = ? WHERE customer_id = ? AND business_id = ?;

-- name: MigrateTimeline :exec
UPDATE crm_customer_timeline SET customer_id = ? WHERE customer_id = ? AND business_id = ?;

-- name: CountLoyaltyTransactions :one
SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?;

-- name: ListLoyaltyTransactions :many
SELECT id, owner_id, business_id, customer_id, type, points, balance_before, balance_after,
       description, reference_id, reference_type, created_at
FROM crm_loyalty_transactions
WHERE business_id = ? AND customer_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountAllLoyaltyTransactions :one
SELECT COUNT(*) FROM crm_loyalty_transactions WHERE business_id = ?;

-- name: ListAllLoyaltyTransactions :many
SELECT id, owner_id, business_id, customer_id, type, points, balance_before, balance_after,
       description, reference_id, reference_type, created_at
FROM crm_loyalty_transactions
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CreateLoyaltyTransaction :exec
INSERT INTO crm_loyalty_transactions
  (owner_id, business_id, customer_id, type, points, balance_before, balance_after,
   description, reference_id, reference_type)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetLoyaltyConfig :one
SELECT id, owner_id, business_id, is_enabled, points_per_currency_unit, currency_per_point,
       bronze_min, silver_min, gold_min, platinum_min,
       bronze_discount_pct, silver_discount_pct, gold_discount_pct, platinum_discount_pct,
       points_expiry_days, updated_at
FROM crm_loyalty_config WHERE business_id = ?;

-- name: UpsertLoyaltyConfig :exec
INSERT INTO crm_loyalty_config
  (owner_id, business_id, is_enabled, points_per_currency_unit, currency_per_point,
   bronze_min, silver_min, gold_min, platinum_min,
   bronze_discount_pct, silver_discount_pct, gold_discount_pct, platinum_discount_pct,
   points_expiry_days)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  is_enabled = VALUES(is_enabled),
  points_per_currency_unit = VALUES(points_per_currency_unit),
  currency_per_point = VALUES(currency_per_point),
  bronze_min = VALUES(bronze_min),
  silver_min = VALUES(silver_min),
  gold_min = VALUES(gold_min),
  platinum_min = VALUES(platinum_min),
  bronze_discount_pct = VALUES(bronze_discount_pct),
  silver_discount_pct = VALUES(silver_discount_pct),
  gold_discount_pct = VALUES(gold_discount_pct),
  platinum_discount_pct = VALUES(platinum_discount_pct),
  points_expiry_days = VALUES(points_expiry_days);

-- name: GetCampaignSendStats :one
SELECT
    COALESCE(SUM(CASE WHEN status = 'bounced' THEN 1 ELSE 0 END), 0) AS bounced_count,
    COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) AS failed_count
FROM crm_campaign_sends WHERE campaign_id = ? AND business_id = ?;

-- name: GetCustomerTimeline :many
SELECT id, owner_id, business_id, customer_id, type, description, metadata, created_at
FROM crm_customer_timeline
WHERE business_id = ? AND customer_id = ?
ORDER BY created_at DESC
LIMIT 100;

-- name: ListCustomerTimelinePage :many
SELECT id, owner_id, business_id, customer_id, type, description, metadata, created_at
FROM crm_customer_timeline
WHERE business_id = sqlc.arg(business_id)
  AND customer_id = sqlc.arg(customer_id)
  AND (
    sqlc.narg(cursor_created_at) IS NULL
    OR created_at < sqlc.narg(cursor_created_at)
    OR (created_at = sqlc.narg(cursor_created_at) AND id < sqlc.arg(cursor_id))
  )
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: InsertTimelineEntry :exec
INSERT INTO crm_customer_timeline (owner_id, business_id, customer_id, type, description, metadata)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListLoyaltyTiers :many
SELECT id, owner_id, business_id, name, min_points, discount_pct, perks, sort_order, created_at, updated_at
FROM crm_loyalty_tiers WHERE business_id = ? ORDER BY sort_order ASC LIMIT 100;

-- name: CreateLoyaltyTier :execresult
INSERT INTO crm_loyalty_tiers (owner_id, business_id, name, min_points, discount_pct, perks, sort_order)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateLoyaltyTier :execresult
UPDATE crm_loyalty_tiers SET name = ?, min_points = ?, discount_pct = ?, perks = ?, sort_order = ?
WHERE id = ? AND business_id = ?;

-- name: GetLoyaltyTierByID :one
SELECT id, owner_id, business_id, name, min_points, discount_pct, perks, sort_order, created_at, updated_at
FROM crm_loyalty_tiers WHERE id = ? AND business_id = ?;

-- name: CountActiveMembersOnTier :one
SELECT COUNT(*) FROM crm_customers
WHERE business_id = ? AND is_active = TRUE AND loyalty_tier = ?;

-- name: DeleteLoyaltyTier :execresult
DELETE FROM crm_loyalty_tiers WHERE id = ? AND business_id = ?;
