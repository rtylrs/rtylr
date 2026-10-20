-- Lead lists share one filter set; each sort has its own keyset query.
-- assigned_to = 0 selects unassigned leads. search matches names, company and
-- email (a LIKE pattern); search_digits matches the normalised phone.

-- name: ListLeadsNewest :many
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('source') IS NULL OR source = sqlc.narg('source'))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL OR created_at < sqlc.narg('cursor_time')
       OR (created_at = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: ListLeadsByFollowUp :many
-- Soonest follow-up first; leads without one come last.
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('source') IS NULL OR source = sqlc.narg('source'))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL
       OR COALESCE(next_follow_up, '9999-12-31') > sqlc.narg('cursor_date')
       OR (COALESCE(next_follow_up, '9999-12-31') = sqlc.narg('cursor_date') AND id > sqlc.narg('cursor_id')))
ORDER BY COALESCE(next_follow_up, '9999-12-31'), id
LIMIT ?;

-- name: ListLeadsByActivity :many
-- Most recently worked first; never-touched leads come last.
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('source') IS NULL OR source = sqlc.narg('source'))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL
       OR COALESCE(last_activity_at, '1970-01-01 00:00:01') < sqlc.narg('cursor_time')
       OR (COALESCE(last_activity_at, '1970-01-01 00:00:01') = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY COALESCE(last_activity_at, '1970-01-01 00:00:01') DESC, id DESC
LIMIT ?;

-- name: ListLeadsByCompany :many
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('source') IS NULL OR source = sqlc.narg('source'))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL OR company_name > sqlc.narg('cursor_text')
       OR (company_name = sqlc.narg('cursor_text') AND id > sqlc.narg('cursor_id')))
ORDER BY company_name, id
LIMIT ?;

-- name: CountLeads :one
SELECT COUNT(*) FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
  AND (sqlc.narg('source') IS NULL OR source = sqlc.narg('source'))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR phone_digits LIKE sqlc.narg('search_digits'));

-- name: CountLeadsByStatus :many
SELECT status, COUNT(*) AS total FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'))
GROUP BY status;

-- name: CountLeadFollowUps :one
-- Follow-ups on leads still being worked: overdue and due on `today`.
SELECT CAST(COALESCE(SUM(next_follow_up < sqlc.arg('today')), 0) AS SIGNED) AS overdue,
       CAST(COALESCE(SUM(next_follow_up = sqlc.arg('today')), 0) AS SIGNED) AS due_today
FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN ('new', 'attempting', 'connected', 'nurture')
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'));

-- name: CountLeadsConvertedSince :one
SELECT COUNT(*) FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND converted_at >= sqlc.arg('since')
  AND (sqlc.narg('assigned_to') IS NULL OR COALESCE(assigned_to, 0) = sqlc.narg('assigned_to'));

-- name: ListLeadSources :many
SELECT DISTINCT source FROM crm_leads
WHERE business_id = ? AND source <> ''
ORDER BY source
LIMIT 200;

-- name: GetLead :one
SELECT * FROM crm_leads WHERE id = ? AND business_id = ?;

-- name: GetLeadForUpdate :one
SELECT * FROM crm_leads WHERE id = ? AND business_id = ? FOR UPDATE;

-- name: CreateLead :execresult
INSERT INTO crm_leads (owner_id, business_id, first_name, last_name, job_title, email, phone, phone_digits,
                       company_name, website, linkedin, industry, size, city, country, source, status,
                       assigned_to, value, billing, notes, import_batch_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateLead :exec
UPDATE crm_leads
SET first_name = ?, last_name = ?, job_title = ?, email = ?, phone = ?, phone_digits = ?,
    company_name = ?, website = ?, linkedin = ?, industry = ?, size = ?, city = ?, country = ?,
    source = ?, status = ?, disqualify_reason = ?, assigned_to = ?, value = ?, billing = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: FillLeadBlanks :exec
-- A re-imported row only fills what the lead is missing; reps' edits win.
UPDATE crm_leads
SET first_name = COALESCE(first_name, sqlc.narg('first_name')),
    last_name = COALESCE(last_name, sqlc.narg('last_name')),
    job_title = COALESCE(job_title, sqlc.narg('job_title')),
    email = COALESCE(email, sqlc.narg('email')),
    phone = COALESCE(phone, sqlc.narg('phone')),
    phone_digits = COALESCE(phone_digits, sqlc.narg('phone_digits')),
    company_name = COALESCE(NULLIF(company_name, ''), sqlc.arg('company_name')),
    website = COALESCE(website, sqlc.narg('website')),
    linkedin = COALESCE(linkedin, sqlc.narg('linkedin')),
    industry = COALESCE(industry, sqlc.narg('industry')),
    size = COALESCE(size, sqlc.narg('size')),
    city = COALESCE(city, sqlc.narg('city')),
    country = COALESCE(country, sqlc.narg('country')),
    value = COALESCE(value, sqlc.narg('value')),
    notes = COALESCE(notes, sqlc.narg('notes'))
WHERE id = sqlc.arg('id') AND business_id = sqlc.arg('business_id');

-- name: RecordLeadWork :exec
UPDATE crm_leads
SET status = ?, last_activity_at = ?, attempt_count = ?
WHERE id = ? AND business_id = ?;

-- name: SetLeadNextFollowUp :exec
UPDATE crm_leads
SET next_follow_up = (
    SELECT MIN(t.due_date) FROM crm_tasks t
    WHERE t.business_id = crm_leads.business_id AND t.lead_id = crm_leads.id AND t.completed_at IS NULL
)
WHERE crm_leads.id = ? AND crm_leads.business_id = ?;

-- name: MarkLeadConverted :exec
UPDATE crm_leads
SET status = 'qualified', disqualify_reason = NULL, company_id = ?, deal_id = ?, converted_at = ?
WHERE id = ? AND business_id = ?;

-- name: LinkLeadCustomer :exec
UPDATE crm_leads SET customer_id = ? WHERE id = ? AND business_id = ?;

-- name: DeleteLead :execresult
DELETE FROM crm_leads WHERE id = ? AND business_id = ? AND status <> 'qualified';

-- name: AssignLeads :execresult
UPDATE crm_leads SET assigned_to = sqlc.narg('assigned_to')
WHERE business_id = sqlc.arg('business_id') AND id IN (sqlc.slice('ids'));

-- name: SetLeadsStatus :execresult
-- Converted leads keep their status; conversion is the only way in or out.
UPDATE crm_leads SET status = sqlc.arg('status'), disqualify_reason = sqlc.narg('disqualify_reason')
WHERE business_id = sqlc.arg('business_id') AND id IN (sqlc.slice('ids')) AND status <> 'qualified';

-- name: DeleteLeads :execresult
DELETE FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND id IN (sqlc.slice('ids')) AND status <> 'qualified';

-- name: FindLeadsByEmail :many
SELECT id, email FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND email IN (sqlc.slice('emails'));

-- name: FindLeadsByPhone :many
SELECT id, phone_digits FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND phone_digits IN (sqlc.slice('phones'));

-- name: FindLeadsByCompany :many
SELECT id, company_name, first_name, last_name FROM crm_leads
WHERE business_id = sqlc.arg('business_id') AND company_name IN (sqlc.slice('companies'));

-- name: DeleteUntouchedBatchLeads :execresult
-- Undo an import: only leads nobody has worked since are removed.
DELETE FROM crm_leads
WHERE business_id = ? AND import_batch_id = ? AND status = 'new'
  AND last_activity_at IS NULL AND next_follow_up IS NULL AND deal_id IS NULL;

-- name: CountBatchLeads :one
SELECT COUNT(*) FROM crm_leads WHERE business_id = ? AND import_batch_id = ?;
