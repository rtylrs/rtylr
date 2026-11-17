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

-- name: RecurringRevenue :one
-- Won monthly and yearly deals are the recurring contracts. Deals have no end
-- date, so every one is active. MRR is the monthly values plus a twelfth of
-- the yearly ones; the month-start figure counts only deals closed before it.
SELECT COUNT(*) AS contracts,
       CAST(COALESCE(SUM(IF(billing = 'monthly', value, 0)) + SUM(IF(billing = 'yearly', value, 0)) / 12, 0) AS DECIMAL(16,2)) AS mrr,
       CAST(COALESCE(SUM(IF(billing = 'monthly' AND closed_at < sqlc.arg('month_start'), value, 0))
           + SUM(IF(billing = 'yearly' AND closed_at < sqlc.arg('month_start'), value, 0)) / 12, 0) AS DECIMAL(16,2)) AS mrr_at_month_start
FROM crm_deals
WHERE business_id = sqlc.arg('business_id') AND status = 'won' AND billing IN ('monthly', 'yearly');
