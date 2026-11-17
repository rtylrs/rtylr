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
-- Recurring revenue comes from CRM contracts. A contract counts when it is
-- signed, has started and has not ended. Its MRR is the mrr field, or, when
-- that was left at zero, its value spread by billing (monthly as is, yearly
-- over twelve). The month-start figure applies the same test on month_start,
-- and also needs the signature to predate it.
SELECT CAST(COALESCE(SUM(IF(
           signed_at IS NOT NULL AND signed_at <= DATE(sqlc.arg('now'))
           AND (start_date IS NULL OR start_date <= DATE(sqlc.arg('now')))
           AND (end_date IS NULL OR end_date >= DATE(sqlc.arg('now'))), 1, 0)), 0) AS SIGNED) AS contracts,
       CAST(COALESCE(SUM(IF(
           signed_at IS NOT NULL AND signed_at <= DATE(sqlc.arg('now'))
           AND (start_date IS NULL OR start_date <= DATE(sqlc.arg('now')))
           AND (end_date IS NULL OR end_date >= DATE(sqlc.arg('now'))),
           IF(mrr > 0, mrr, CASE billing WHEN 'monthly' THEN value WHEN 'yearly' THEN value / 12 ELSE 0 END), 0)), 0) AS DECIMAL(16,2)) AS mrr,
       CAST(COALESCE(SUM(IF(
           signed_at IS NOT NULL AND signed_at < DATE(sqlc.arg('month_start'))
           AND (start_date IS NULL OR start_date < DATE(sqlc.arg('month_start')))
           AND (end_date IS NULL OR end_date >= DATE(sqlc.arg('month_start'))),
           IF(mrr > 0, mrr, CASE billing WHEN 'monthly' THEN value WHEN 'yearly' THEN value / 12 ELSE 0 END), 0)), 0) AS DECIMAL(16,2)) AS mrr_at_month_start
FROM crm_contracts
WHERE business_id = sqlc.arg('business_id') AND billing IN ('monthly', 'yearly');
