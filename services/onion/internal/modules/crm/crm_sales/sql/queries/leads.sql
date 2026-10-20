-- Lead lists share one filter block, kept identical in every query below
-- (leads_filter_test.go checks). Each sort has its own keyset query.
--
-- A set filter applies when its filter_* flag is 1; sqlc turns an empty slice
-- into IN (NULL). filter_reps with unassigned = 1 also matches leads without a
-- rep. Flags are 0 or 1 and has_* are 'yes' or 'no' (sqlc types a yes/no on a
-- date column as a date, so those use flags). *_before bounds are exclusive.
-- search matches names, company, email and job title (a LIKE pattern);
-- search_digits matches the normalised phone.

-- name: ListLeadsNewest :many
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (CAST(sqlc.arg('filter_reps') AS UNSIGNED) = 0 OR assigned_to IN (sqlc.slice('reps'))
       OR (CAST(sqlc.arg('unassigned') AS UNSIGNED) = 1 AND assigned_to IS NULL))
  AND (CAST(sqlc.arg('filter_sources') AS UNSIGNED) = 0 OR source IN (sqlc.slice('sources')))
  AND (CAST(sqlc.arg('filter_industries') AS UNSIGNED) = 0 OR industry IN (sqlc.slice('industries')))
  AND (CAST(sqlc.arg('filter_sizes') AS UNSIGNED) = 0 OR size IN (sqlc.slice('sizes')))
  AND (CAST(sqlc.arg('filter_countries') AS UNSIGNED) = 0 OR country IN (sqlc.slice('countries')))
  AND (CAST(sqlc.arg('filter_cities') AS UNSIGNED) = 0 OR city IN (sqlc.slice('cities')))
  AND (CAST(sqlc.arg('filter_billing') AS UNSIGNED) = 0 OR billing IN (sqlc.slice('billing')))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (CAST(sqlc.arg('no_follow_up') AS UNSIGNED) = 0 OR next_follow_up IS NULL)
  AND (sqlc.narg('has_phone') IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = sqlc.narg('has_phone'))
  AND (sqlc.narg('has_email') IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = sqlc.narg('has_email'))
  AND (sqlc.narg('has_website') IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = sqlc.narg('has_website'))
  AND (sqlc.narg('has_linkedin') IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = sqlc.narg('has_linkedin'))
  AND (sqlc.narg('has_company') IS NULL OR IF(company_name = '', 'no', 'yes') = sqlc.narg('has_company'))
  AND (sqlc.narg('has_value') IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = sqlc.narg('has_value'))
  AND (CAST(sqlc.arg('never_contacted') AS UNSIGNED) = 0 OR last_activity_at IS NULL)
  AND (sqlc.narg('attempts_min') IS NULL OR attempt_count >= sqlc.narg('attempts_min'))
  AND (sqlc.narg('attempts_max') IS NULL OR attempt_count <= sqlc.narg('attempts_max'))
  AND (sqlc.narg('value_min') IS NULL OR value >= sqlc.narg('value_min'))
  AND (sqlc.narg('value_max') IS NULL OR value <= sqlc.narg('value_max'))
  AND (sqlc.narg('created_from') IS NULL OR created_at >= sqlc.narg('created_from'))
  AND (sqlc.narg('created_before') IS NULL OR created_at < sqlc.narg('created_before'))
  AND (sqlc.narg('active_from') IS NULL OR last_activity_at >= sqlc.narg('active_from'))
  AND (sqlc.narg('active_before') IS NULL OR last_activity_at < sqlc.narg('active_before'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR job_title LIKE sqlc.narg('search')
       OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL OR created_at < sqlc.narg('cursor_time')
       OR (created_at = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: ListLeadsByFollowUp :many
-- Soonest follow-up first; leads without one come last.
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (CAST(sqlc.arg('filter_reps') AS UNSIGNED) = 0 OR assigned_to IN (sqlc.slice('reps'))
       OR (CAST(sqlc.arg('unassigned') AS UNSIGNED) = 1 AND assigned_to IS NULL))
  AND (CAST(sqlc.arg('filter_sources') AS UNSIGNED) = 0 OR source IN (sqlc.slice('sources')))
  AND (CAST(sqlc.arg('filter_industries') AS UNSIGNED) = 0 OR industry IN (sqlc.slice('industries')))
  AND (CAST(sqlc.arg('filter_sizes') AS UNSIGNED) = 0 OR size IN (sqlc.slice('sizes')))
  AND (CAST(sqlc.arg('filter_countries') AS UNSIGNED) = 0 OR country IN (sqlc.slice('countries')))
  AND (CAST(sqlc.arg('filter_cities') AS UNSIGNED) = 0 OR city IN (sqlc.slice('cities')))
  AND (CAST(sqlc.arg('filter_billing') AS UNSIGNED) = 0 OR billing IN (sqlc.slice('billing')))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (CAST(sqlc.arg('no_follow_up') AS UNSIGNED) = 0 OR next_follow_up IS NULL)
  AND (sqlc.narg('has_phone') IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = sqlc.narg('has_phone'))
  AND (sqlc.narg('has_email') IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = sqlc.narg('has_email'))
  AND (sqlc.narg('has_website') IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = sqlc.narg('has_website'))
  AND (sqlc.narg('has_linkedin') IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = sqlc.narg('has_linkedin'))
  AND (sqlc.narg('has_company') IS NULL OR IF(company_name = '', 'no', 'yes') = sqlc.narg('has_company'))
  AND (sqlc.narg('has_value') IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = sqlc.narg('has_value'))
  AND (CAST(sqlc.arg('never_contacted') AS UNSIGNED) = 0 OR last_activity_at IS NULL)
  AND (sqlc.narg('attempts_min') IS NULL OR attempt_count >= sqlc.narg('attempts_min'))
  AND (sqlc.narg('attempts_max') IS NULL OR attempt_count <= sqlc.narg('attempts_max'))
  AND (sqlc.narg('value_min') IS NULL OR value >= sqlc.narg('value_min'))
  AND (sqlc.narg('value_max') IS NULL OR value <= sqlc.narg('value_max'))
  AND (sqlc.narg('created_from') IS NULL OR created_at >= sqlc.narg('created_from'))
  AND (sqlc.narg('created_before') IS NULL OR created_at < sqlc.narg('created_before'))
  AND (sqlc.narg('active_from') IS NULL OR last_activity_at >= sqlc.narg('active_from'))
  AND (sqlc.narg('active_before') IS NULL OR last_activity_at < sqlc.narg('active_before'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR job_title LIKE sqlc.narg('search')
       OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL
       OR COALESCE(next_follow_up, '9999-12-31') > sqlc.narg('cursor_time')
       OR (COALESCE(next_follow_up, '9999-12-31') = sqlc.narg('cursor_time') AND id > sqlc.narg('cursor_id')))
ORDER BY COALESCE(next_follow_up, '9999-12-31'), id
LIMIT ?;

-- name: ListLeadsByActivity :many
-- Most recently worked first; never-touched leads come last.
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (CAST(sqlc.arg('filter_reps') AS UNSIGNED) = 0 OR assigned_to IN (sqlc.slice('reps'))
       OR (CAST(sqlc.arg('unassigned') AS UNSIGNED) = 1 AND assigned_to IS NULL))
  AND (CAST(sqlc.arg('filter_sources') AS UNSIGNED) = 0 OR source IN (sqlc.slice('sources')))
  AND (CAST(sqlc.arg('filter_industries') AS UNSIGNED) = 0 OR industry IN (sqlc.slice('industries')))
  AND (CAST(sqlc.arg('filter_sizes') AS UNSIGNED) = 0 OR size IN (sqlc.slice('sizes')))
  AND (CAST(sqlc.arg('filter_countries') AS UNSIGNED) = 0 OR country IN (sqlc.slice('countries')))
  AND (CAST(sqlc.arg('filter_cities') AS UNSIGNED) = 0 OR city IN (sqlc.slice('cities')))
  AND (CAST(sqlc.arg('filter_billing') AS UNSIGNED) = 0 OR billing IN (sqlc.slice('billing')))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (CAST(sqlc.arg('no_follow_up') AS UNSIGNED) = 0 OR next_follow_up IS NULL)
  AND (sqlc.narg('has_phone') IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = sqlc.narg('has_phone'))
  AND (sqlc.narg('has_email') IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = sqlc.narg('has_email'))
  AND (sqlc.narg('has_website') IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = sqlc.narg('has_website'))
  AND (sqlc.narg('has_linkedin') IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = sqlc.narg('has_linkedin'))
  AND (sqlc.narg('has_company') IS NULL OR IF(company_name = '', 'no', 'yes') = sqlc.narg('has_company'))
  AND (sqlc.narg('has_value') IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = sqlc.narg('has_value'))
  AND (CAST(sqlc.arg('never_contacted') AS UNSIGNED) = 0 OR last_activity_at IS NULL)
  AND (sqlc.narg('attempts_min') IS NULL OR attempt_count >= sqlc.narg('attempts_min'))
  AND (sqlc.narg('attempts_max') IS NULL OR attempt_count <= sqlc.narg('attempts_max'))
  AND (sqlc.narg('value_min') IS NULL OR value >= sqlc.narg('value_min'))
  AND (sqlc.narg('value_max') IS NULL OR value <= sqlc.narg('value_max'))
  AND (sqlc.narg('created_from') IS NULL OR created_at >= sqlc.narg('created_from'))
  AND (sqlc.narg('created_before') IS NULL OR created_at < sqlc.narg('created_before'))
  AND (sqlc.narg('active_from') IS NULL OR last_activity_at >= sqlc.narg('active_from'))
  AND (sqlc.narg('active_before') IS NULL OR last_activity_at < sqlc.narg('active_before'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR job_title LIKE sqlc.narg('search')
       OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL
       OR COALESCE(last_activity_at, '1970-01-01 00:00:01') < sqlc.narg('cursor_time')
       OR (COALESCE(last_activity_at, '1970-01-01 00:00:01') = sqlc.narg('cursor_time') AND id < sqlc.narg('cursor_id')))
ORDER BY COALESCE(last_activity_at, '1970-01-01 00:00:01') DESC, id DESC
LIMIT ?;

-- name: ListLeadsByCompany :many
SELECT * FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND status IN (sqlc.slice('statuses'))
  AND (CAST(sqlc.arg('filter_reps') AS UNSIGNED) = 0 OR assigned_to IN (sqlc.slice('reps'))
       OR (CAST(sqlc.arg('unassigned') AS UNSIGNED) = 1 AND assigned_to IS NULL))
  AND (CAST(sqlc.arg('filter_sources') AS UNSIGNED) = 0 OR source IN (sqlc.slice('sources')))
  AND (CAST(sqlc.arg('filter_industries') AS UNSIGNED) = 0 OR industry IN (sqlc.slice('industries')))
  AND (CAST(sqlc.arg('filter_sizes') AS UNSIGNED) = 0 OR size IN (sqlc.slice('sizes')))
  AND (CAST(sqlc.arg('filter_countries') AS UNSIGNED) = 0 OR country IN (sqlc.slice('countries')))
  AND (CAST(sqlc.arg('filter_cities') AS UNSIGNED) = 0 OR city IN (sqlc.slice('cities')))
  AND (CAST(sqlc.arg('filter_billing') AS UNSIGNED) = 0 OR billing IN (sqlc.slice('billing')))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (CAST(sqlc.arg('no_follow_up') AS UNSIGNED) = 0 OR next_follow_up IS NULL)
  AND (sqlc.narg('has_phone') IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = sqlc.narg('has_phone'))
  AND (sqlc.narg('has_email') IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = sqlc.narg('has_email'))
  AND (sqlc.narg('has_website') IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = sqlc.narg('has_website'))
  AND (sqlc.narg('has_linkedin') IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = sqlc.narg('has_linkedin'))
  AND (sqlc.narg('has_company') IS NULL OR IF(company_name = '', 'no', 'yes') = sqlc.narg('has_company'))
  AND (sqlc.narg('has_value') IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = sqlc.narg('has_value'))
  AND (CAST(sqlc.arg('never_contacted') AS UNSIGNED) = 0 OR last_activity_at IS NULL)
  AND (sqlc.narg('attempts_min') IS NULL OR attempt_count >= sqlc.narg('attempts_min'))
  AND (sqlc.narg('attempts_max') IS NULL OR attempt_count <= sqlc.narg('attempts_max'))
  AND (sqlc.narg('value_min') IS NULL OR value >= sqlc.narg('value_min'))
  AND (sqlc.narg('value_max') IS NULL OR value <= sqlc.narg('value_max'))
  AND (sqlc.narg('created_from') IS NULL OR created_at >= sqlc.narg('created_from'))
  AND (sqlc.narg('created_before') IS NULL OR created_at < sqlc.narg('created_before'))
  AND (sqlc.narg('active_from') IS NULL OR last_activity_at >= sqlc.narg('active_from'))
  AND (sqlc.narg('active_before') IS NULL OR last_activity_at < sqlc.narg('active_before'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR job_title LIKE sqlc.narg('search')
       OR phone_digits LIKE sqlc.narg('search_digits'))
  AND (sqlc.narg('cursor_id') IS NULL OR company_name > sqlc.narg('cursor_text')
       OR (company_name = sqlc.narg('cursor_text') AND id > sqlc.narg('cursor_id')))
ORDER BY company_name, id
LIMIT ?;

-- name: CountLeadsByStatus :many
-- Counts the filter's leads per status (the status filter aside, for the
-- tabs), with the follow-ups due on or before today and the leads converted
-- since month_start among them.
SELECT status, COUNT(*) AS total,
       CAST(COALESCE(SUM(next_follow_up < sqlc.arg('today')), 0) AS SIGNED) AS overdue,
       CAST(COALESCE(SUM(next_follow_up = sqlc.arg('today')), 0) AS SIGNED) AS due_today,
       CAST(COALESCE(SUM(converted_at >= sqlc.arg('month_start')), 0) AS SIGNED) AS converted
FROM crm_leads
WHERE business_id = sqlc.arg('business_id')
  AND (CAST(sqlc.arg('filter_reps') AS UNSIGNED) = 0 OR assigned_to IN (sqlc.slice('reps'))
       OR (CAST(sqlc.arg('unassigned') AS UNSIGNED) = 1 AND assigned_to IS NULL))
  AND (CAST(sqlc.arg('filter_sources') AS UNSIGNED) = 0 OR source IN (sqlc.slice('sources')))
  AND (CAST(sqlc.arg('filter_industries') AS UNSIGNED) = 0 OR industry IN (sqlc.slice('industries')))
  AND (CAST(sqlc.arg('filter_sizes') AS UNSIGNED) = 0 OR size IN (sqlc.slice('sizes')))
  AND (CAST(sqlc.arg('filter_countries') AS UNSIGNED) = 0 OR country IN (sqlc.slice('countries')))
  AND (CAST(sqlc.arg('filter_cities') AS UNSIGNED) = 0 OR city IN (sqlc.slice('cities')))
  AND (CAST(sqlc.arg('filter_billing') AS UNSIGNED) = 0 OR billing IN (sqlc.slice('billing')))
  AND (sqlc.narg('import_batch_id') IS NULL OR import_batch_id = sqlc.narg('import_batch_id'))
  AND (sqlc.narg('follow_up_from') IS NULL OR next_follow_up >= sqlc.narg('follow_up_from'))
  AND (sqlc.narg('follow_up_to') IS NULL OR next_follow_up <= sqlc.narg('follow_up_to'))
  AND (CAST(sqlc.arg('no_follow_up') AS UNSIGNED) = 0 OR next_follow_up IS NULL)
  AND (sqlc.narg('has_phone') IS NULL OR IF(COALESCE(phone_digits, '') = '', 'no', 'yes') = sqlc.narg('has_phone'))
  AND (sqlc.narg('has_email') IS NULL OR IF(COALESCE(email, '') = '', 'no', 'yes') = sqlc.narg('has_email'))
  AND (sqlc.narg('has_website') IS NULL OR IF(COALESCE(website, '') = '', 'no', 'yes') = sqlc.narg('has_website'))
  AND (sqlc.narg('has_linkedin') IS NULL OR IF(COALESCE(linkedin, '') = '', 'no', 'yes') = sqlc.narg('has_linkedin'))
  AND (sqlc.narg('has_company') IS NULL OR IF(company_name = '', 'no', 'yes') = sqlc.narg('has_company'))
  AND (sqlc.narg('has_value') IS NULL OR IF(COALESCE(value, 0) = 0, 'no', 'yes') = sqlc.narg('has_value'))
  AND (CAST(sqlc.arg('never_contacted') AS UNSIGNED) = 0 OR last_activity_at IS NULL)
  AND (sqlc.narg('attempts_min') IS NULL OR attempt_count >= sqlc.narg('attempts_min'))
  AND (sqlc.narg('attempts_max') IS NULL OR attempt_count <= sqlc.narg('attempts_max'))
  AND (sqlc.narg('value_min') IS NULL OR value >= sqlc.narg('value_min'))
  AND (sqlc.narg('value_max') IS NULL OR value <= sqlc.narg('value_max'))
  AND (sqlc.narg('created_from') IS NULL OR created_at >= sqlc.narg('created_from'))
  AND (sqlc.narg('created_before') IS NULL OR created_at < sqlc.narg('created_before'))
  AND (sqlc.narg('active_from') IS NULL OR last_activity_at >= sqlc.narg('active_from'))
  AND (sqlc.narg('active_before') IS NULL OR last_activity_at < sqlc.narg('active_before'))
  AND (sqlc.narg('search') IS NULL OR company_name LIKE sqlc.narg('search')
       OR CONCAT_WS(' ', first_name, last_name) LIKE sqlc.narg('search')
       OR email LIKE sqlc.narg('search') OR job_title LIKE sqlc.narg('search')
       OR phone_digits LIKE sqlc.narg('search_digits'))
GROUP BY status;

-- Distinct values leads have, for the filter pickers.

-- name: ListLeadSources :many
SELECT DISTINCT source FROM crm_leads
WHERE business_id = ? AND source <> ''
ORDER BY source
LIMIT 500;

-- name: ListLeadIndustries :many
SELECT DISTINCT industry FROM crm_leads
WHERE business_id = ? AND industry <> ''
ORDER BY industry
LIMIT 500;

-- name: ListLeadSizes :many
SELECT DISTINCT size FROM crm_leads
WHERE business_id = ? AND size <> ''
ORDER BY size
LIMIT 500;

-- name: ListLeadCountries :many
SELECT DISTINCT country FROM crm_leads
WHERE business_id = ? AND country <> ''
ORDER BY country
LIMIT 500;

-- name: ListLeadCities :many
SELECT DISTINCT city FROM crm_leads
WHERE business_id = ? AND city <> ''
ORDER BY city
LIMIT 500;

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
