-- Sales activity over a window, for the CRM dashboard. Days are the caller's:
-- timestamps shift by offset_minutes (the caller's offset from UTC) before the
-- day is taken, as local_day (YYYY-MM-DD). `rep` narrows to one person: calls
-- by who logged them, leads and deals by who they are assigned to (0 for
-- unassigned).

-- name: CallsByDay :many
SELECT DATE_FORMAT(DATE_ADD(occurred_at, INTERVAL CAST(sqlc.arg('offset_minutes') AS SIGNED) MINUTE), '%Y-%m-%d') AS local_day,
       COUNT(*) AS calls,
       CAST(COALESCE(SUM(outcome IN ('connected', 'callback', 'meeting_booked', 'not_interested')), 0) AS SIGNED) AS reached
FROM crm_touchpoints
WHERE business_id = sqlc.arg('business_id') AND type = 'call' AND occurred_at >= sqlc.arg('since')
  AND (sqlc.narg('rep') IS NULL OR created_by = sqlc.narg('rep'))
GROUP BY local_day
ORDER BY local_day;

-- name: LeadsAddedByDay :many
SELECT DATE_FORMAT(DATE_ADD(created_at, INTERVAL CAST(sqlc.arg('offset_minutes') AS SIGNED) MINUTE), '%Y-%m-%d') AS local_day,
       COUNT(*) AS leads
FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND created_at >= sqlc.arg('since')
  AND (sqlc.narg('rep') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('rep'))
GROUP BY local_day
ORDER BY local_day;

-- name: DealsWonByDay :many
SELECT DATE_FORMAT(DATE_ADD(closed_at, INTERVAL CAST(sqlc.arg('offset_minutes') AS SIGNED) MINUTE), '%Y-%m-%d') AS local_day,
       COUNT(*) AS won
FROM crm_deals
WHERE business_id = sqlc.arg('business_id') AND status = 'won' AND closed_at >= sqlc.arg('since')
  AND (sqlc.narg('rep') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('rep'))
GROUP BY local_day
ORDER BY local_day;

-- name: CallsByRep :many
SELECT created_by AS rep_id,
       COUNT(*) AS calls,
       CAST(COALESCE(SUM(outcome IN ('connected', 'callback', 'meeting_booked', 'not_interested')), 0) AS SIGNED) AS reached
FROM crm_touchpoints
WHERE business_id = sqlc.arg('business_id') AND type = 'call' AND occurred_at >= sqlc.arg('since')
  AND (sqlc.narg('rep') IS NULL OR created_by = sqlc.narg('rep'))
GROUP BY created_by;

-- name: LeadsQualifiedByRep :many
SELECT CAST(COALESCE(assigned_to, 0) AS SIGNED) AS rep_id, COUNT(*) AS qualified
FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND converted_at >= sqlc.arg('since')
  AND (sqlc.narg('rep') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('rep'))
GROUP BY rep_id;

-- name: DealOutcomesByRep :many
SELECT CAST(COALESCE(assigned_to, 0) AS SIGNED) AS rep_id,
       CAST(COALESCE(SUM(status = 'won'), 0) AS SIGNED) AS won,
       CAST(COALESCE(SUM(status = 'lost'), 0) AS SIGNED) AS lost,
       CAST(COALESCE(SUM(IF(status = 'won', CASE billing WHEN 'monthly' THEN value * 12 ELSE value END, 0)), 0) AS DECIMAL(16,2)) AS won_amount
FROM crm_deals
WHERE business_id = sqlc.arg('business_id') AND status IN ('won', 'lost') AND closed_at >= sqlc.arg('since')
  AND (sqlc.narg('rep') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('rep'))
GROUP BY rep_id;
