-- Execution side of the automations module. The dispatcher
-- (internal/platform/eventbus) writes one automation_runs row per automation
-- that subscribes to an event; these statements drain that queue.

-- cross-schema read from shared — automation_runs is the platform fan-out queue.
-- Ordered by due time so a run backed off to later does not overtake one that
-- is ready now. 'failed' is included because it means retryable; 'dead' has
-- exhausted its attempts and is deliberately never picked up again.
-- name: ListDueAutomationRuns :many
SELECT id, business_id, automation_id, event_id, attempts, max_attempts
FROM automation_runs
WHERE status IN ('pending', 'failed')
  AND next_attempt_at <= NOW()
ORDER BY next_attempt_at ASC, id ASC
LIMIT ?;

-- Optimistic claim. The status predicate is what makes it safe for several
-- workers to read the same due list: exactly one UPDATE affects a row, and a
-- caller that affects zero rows lost the race and skips.
--
-- attempts increments here rather than on failure, so a run that crashes the
-- worker still counts against max_attempts instead of retrying forever.
-- name: ClaimAutomationRun :execresult
UPDATE automation_runs
SET status = 'running', attempts = attempts + 1, updated_at = NOW()
WHERE id = ? AND status IN ('pending', 'failed');

-- name: MarkAutomationRunSucceeded :exec
UPDATE automation_runs
SET status = 'succeeded', completed_at = NOW(), last_error = NULL, updated_at = NOW()
WHERE id = ?;

-- 'dead' once attempts reach the cap, so an automation that can never succeed
-- stops consuming the queue and stays visible for someone to look at, rather
-- than disappearing or retrying forever.
-- name: MarkAutomationRunFailed :exec
UPDATE automation_runs
SET status = IF(attempts >= max_attempts, 'dead', 'failed'),
    next_attempt_at = DATE_ADD(NOW(), INTERVAL sqlc.arg('retry_after_seconds') SECOND),
    last_error = sqlc.arg('last_error'),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- A permanent failure skips the retries entirely. Retrying something that
-- cannot succeed, such as an action this build does not implement, only delays
-- the operator noticing.
-- name: MarkAutomationRunDead :exec
UPDATE automation_runs
SET status = 'dead', last_error = ?, updated_at = NOW()
WHERE id = ?;

-- Scoped by business_id as well as id: the run carries both, and matching on
-- id alone would let a corrupted run reach another tenant's automation.
-- name: GetAutomationForRun :one
SELECT id, trigger_type, actions, is_active
FROM crm_automations
WHERE id = ? AND business_id = ?;

-- cross-schema read from shared — the event the run was created from.
-- name: GetEventForRun :one
SELECT id, event_type, subject_type, subject_id, payload
FROM platform_events
WHERE id = ? AND business_id = ?;

-- The counters #834 records as dead. Written once per successful execution.
-- name: RecordAutomationExecution :exec
UPDATE crm_automations
SET run_count = run_count + 1, last_run_at = NOW()
WHERE id = ? AND business_id = ?;

-- cross-schema read from crm — the recipient for a send_email action.
-- The customer id comes from the event payload, so it is checked against the
-- run's business rather than trusted.
-- name: GetCustomerContactForRun :one
SELECT id,
       COALESCE(email, '') AS email,
       TRIM(CONCAT(first_name, ' ', COALESCE(last_name, ''))) AS full_name
FROM crm_customers
WHERE id = ? AND business_id = ?;
