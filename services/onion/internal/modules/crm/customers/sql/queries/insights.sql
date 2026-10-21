-- Reads served to Insights reports, so Insights never queries CRM tables.

-- name: InsightsCountCustomers :one
SELECT COUNT(*) AS count FROM crm_customers WHERE business_id = ?;

-- Customers that existed by the end of the calendar day as_of.
-- name: InsightsCountCustomersAsOf :one
SELECT COUNT(*) AS count
FROM crm_customers
WHERE business_id = sqlc.arg(business_id)
  AND created_at < DATE_ADD(sqlc.arg(as_of), INTERVAL 1 DAY);

-- Customers created during the calendar days date_from to date_to.
-- name: InsightsCountNewCustomers :one
SELECT COUNT(*) AS count
FROM crm_customers
WHERE business_id = sqlc.arg(business_id)
  AND created_at >= sqlc.arg(date_from)
  AND created_at < DATE_ADD(sqlc.arg(date_to), INTERVAL 1 DAY);

-- name: InsightsCustomerContacts :many
SELECT id,
       COALESCE(first_name, '') AS first_name,
       COALESCE(last_name, '') AS last_name,
       COALESCE(email, '') AS email
FROM crm_customers
WHERE business_id = sqlc.arg(business_id)
  AND id IN (sqlc.slice(ids));

-- name: InsightsCRMLastChange :one
SELECT CAST(COALESCE(MAX(updated_at), '1000-01-01') AS DATETIME) AS last_change
FROM crm_customers
WHERE business_id = ?;
