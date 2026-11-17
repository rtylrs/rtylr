-- Deal totals for the Insights business overview. Amounts are first-year
-- values, the same as the CRM's own pipeline totals.

-- name: OpenDealPipeline :one
SELECT COUNT(*) AS deals,
       CAST(COALESCE(SUM(CASE billing WHEN 'monthly' THEN value * 12 ELSE value END), 0) AS DECIMAL(16,2)) AS amount
FROM crm_deals
WHERE business_id = sqlc.arg('business_id') AND status = 'open';

-- name: WonDealsBetween :one
SELECT COUNT(*) AS deals,
       CAST(COALESCE(SUM(CASE billing WHEN 'monthly' THEN value * 12 ELSE value END), 0) AS DECIMAL(16,2)) AS amount
FROM crm_deals
WHERE business_id = sqlc.arg('business_id') AND status = 'won'
  AND closed_at >= sqlc.arg('date_from') AND closed_at <= sqlc.arg('date_to');
