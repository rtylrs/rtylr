-- A deal's amount for totals is its first-year value: monthly deals count
-- twelve months, one-time and yearly deals their value. assigned_to = 0
-- selects unassigned deals.

-- name: ListStageDeals :many
-- One board column, newest first.
SELECT * FROM crm_deals
WHERE business_id = sqlc.arg('business_id') AND pipeline_id = sqlc.arg('pipeline_id')
  AND stage_id = sqlc.arg('stage_id')
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('search') IS NULL OR title LIKE sqlc.narg('search')
       OR company_name LIKE sqlc.narg('search') OR contact_name LIKE sqlc.narg('search'))
  AND (sqlc.narg('cursor_id') IS NULL OR created_at < sqlc.narg('cursor_time')
       OR (created_at = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: ListDealsNewest :many
SELECT * FROM crm_deals
WHERE business_id = sqlc.arg('business_id')
  AND (sqlc.narg('pipeline_id') IS NULL OR pipeline_id = sqlc.narg('pipeline_id'))
  AND (sqlc.narg('stage_id') IS NULL OR stage_id = sqlc.narg('stage_id'))
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('search') IS NULL OR title LIKE sqlc.narg('search')
       OR company_name LIKE sqlc.narg('search') OR contact_name LIKE sqlc.narg('search'))
  AND (sqlc.narg('cursor_id') IS NULL OR created_at < sqlc.narg('cursor_time')
       OR (created_at = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: ListDealsByCloseDate :many
-- Soonest expected close first; deals without a date come last.
SELECT * FROM crm_deals
WHERE business_id = sqlc.arg('business_id')
  AND (sqlc.narg('pipeline_id') IS NULL OR pipeline_id = sqlc.narg('pipeline_id'))
  AND (sqlc.narg('stage_id') IS NULL OR stage_id = sqlc.narg('stage_id'))
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('search') IS NULL OR title LIKE sqlc.narg('search')
       OR company_name LIKE sqlc.narg('search') OR contact_name LIKE sqlc.narg('search'))
  AND (sqlc.narg('cursor_id') IS NULL
       OR COALESCE(expected_close_date, '9999-12-31') > sqlc.narg('cursor_date')
       OR (COALESCE(expected_close_date, '9999-12-31') = sqlc.narg('cursor_date') AND id > sqlc.narg('cursor_id')))
ORDER BY COALESCE(expected_close_date, '9999-12-31'), id
LIMIT ?;

-- name: ListDealsByValue :many
SELECT * FROM crm_deals
WHERE business_id = sqlc.arg('business_id')
  AND (sqlc.narg('pipeline_id') IS NULL OR pipeline_id = sqlc.narg('pipeline_id'))
  AND (sqlc.narg('stage_id') IS NULL OR stage_id = sqlc.narg('stage_id'))
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('search') IS NULL OR title LIKE sqlc.narg('search')
       OR company_name LIKE sqlc.narg('search') OR contact_name LIKE sqlc.narg('search'))
  AND (sqlc.narg('cursor_id') IS NULL
       OR value * IF(billing = 'monthly', 12, 1) < sqlc.narg('cursor_amount')
       OR (value * IF(billing = 'monthly', 12, 1) = sqlc.narg('cursor_amount') AND id < sqlc.narg('cursor_id')))
ORDER BY value * IF(billing = 'monthly', 12, 1) DESC, id DESC
LIMIT ?;

-- name: CountDeals :one
SELECT COUNT(*) FROM crm_deals
WHERE business_id = sqlc.arg('business_id')
  AND (sqlc.narg('pipeline_id') IS NULL OR pipeline_id = sqlc.narg('pipeline_id'))
  AND (sqlc.narg('stage_id') IS NULL OR stage_id = sqlc.narg('stage_id'))
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('search') IS NULL OR title LIKE sqlc.narg('search')
       OR company_name LIKE sqlc.narg('search') OR contact_name LIKE sqlc.narg('search'));

-- name: DealStageTotals :many
SELECT stage_id, COUNT(*) AS deals,
       CAST(COALESCE(SUM(CASE billing WHEN 'monthly' THEN value * 12 ELSE value END), 0) AS DECIMAL(16,2)) AS amount
FROM crm_deals
WHERE business_id = sqlc.arg('business_id') AND pipeline_id = sqlc.arg('pipeline_id')
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('search') IS NULL OR title LIKE sqlc.narg('search')
       OR company_name LIKE sqlc.narg('search') OR contact_name LIKE sqlc.narg('search'))
GROUP BY stage_id;

-- name: DealOutcomesSince :one
-- Deals closed on or after `since`, for win rate, won value and cycle time.
SELECT CAST(COALESCE(SUM(status = 'won'), 0) AS SIGNED) AS won,
       CAST(COALESCE(SUM(status = 'lost'), 0) AS SIGNED) AS lost,
       CAST(COALESCE(SUM(IF(status = 'won', CASE billing WHEN 'monthly' THEN value * 12 ELSE value END, 0)), 0) AS DECIMAL(16,2)) AS won_amount,
       CAST(COALESCE(AVG(IF(status = 'won', DATEDIFF(closed_at, created_at), NULL)), 0) AS DECIMAL(10,1)) AS average_days_to_win
FROM crm_deals
WHERE business_id = sqlc.arg('business_id')
  AND (sqlc.narg('pipeline_id') IS NULL OR pipeline_id = sqlc.narg('pipeline_id'))
  AND status IN ('won', 'lost') AND closed_at >= sqlc.arg('since')
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'));

-- name: GetDeal :one
SELECT * FROM crm_deals WHERE id = ? AND business_id = ?;

-- name: GetDealForUpdate :one
SELECT * FROM crm_deals WHERE id = ? AND business_id = ? FOR UPDATE;

-- name: CreateDeal :execresult
INSERT INTO crm_deals (business_id, owner_id, pipeline_id, stage_id, title, lead_id, customer_id, company_id,
                       contact_name, company_name, value, billing, expected_close_date, assigned_to, source,
                       status, closed_at, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateDeal :exec
UPDATE crm_deals
SET title = ?, contact_name = ?, company_name = ?, value = ?, billing = ?, expected_close_date = ?,
    assigned_to = ?, source = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: MoveDeal :exec
UPDATE crm_deals
SET pipeline_id = ?, stage_id = ?, status = ?, closed_at = ?, lost_reason = ?, stage_changed_at = ?
WHERE id = ? AND business_id = ?;

-- name: RecordDealWork :exec
UPDATE crm_deals SET last_activity_at = ? WHERE id = ? AND business_id = ?;

-- name: SetDealNextFollowUp :exec
UPDATE crm_deals
SET next_follow_up = (
    SELECT MIN(t.due_date) FROM crm_tasks t
    WHERE t.business_id = crm_deals.business_id AND t.deal_id = crm_deals.id AND t.completed_at IS NULL
)
WHERE crm_deals.id = ? AND crm_deals.business_id = ?;

-- name: LinkLeadDealsCustomer :exec
UPDATE crm_deals SET customer_id = ?
WHERE business_id = ? AND lead_id = ? AND customer_id IS NULL;

-- name: DeleteDeal :execresult
DELETE FROM crm_deals WHERE id = ? AND business_id = ?;

-- name: CountPipelineDeals :one
SELECT COUNT(*) FROM crm_deals WHERE business_id = ? AND pipeline_id = ?;

-- name: CountStageDeals :one
SELECT COUNT(*) FROM crm_deals WHERE business_id = ? AND stage_id = ?;

-- name: DetachLeadFromDeal :exec
UPDATE crm_leads SET deal_id = NULL, status = 'connected', converted_at = NULL
WHERE business_id = ? AND deal_id = ?;
