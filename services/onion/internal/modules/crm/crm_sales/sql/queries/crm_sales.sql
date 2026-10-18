-- name: ListLeads :many
SELECT l.id, l.owner_id, l.business_id, l.company_name, COALESCE(l.industry,'') AS industry,
       COALESCE(l.size,'') AS size, l.source, l.stage, l.assigned_to, l.mrr_potential,
       COALESCE(l.notes,'') AS notes, l.created_at, l.updated_at,
       -- The follow-up date on the latest touchpoint; logging a new touchpoint
       -- replaces it, or clears it when the new one sets no follow-up.
       CAST(COALESCE(DATE_FORMAT((SELECT t.next_follow_up FROM crm_touchpoints t
             WHERE t.lead_id = l.id AND t.business_id = l.business_id
             ORDER BY t.created_at DESC, t.id DESC LIMIT 1), '%Y-%m-%d'), '') AS CHAR) AS next_follow_up
FROM crm_leads l
WHERE l.business_id = ?
ORDER BY l.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountLeads :one
SELECT COUNT(*) FROM crm_leads WHERE business_id = ?;

-- name: GetLead :one
SELECT l.id, l.owner_id, l.business_id, l.company_name, COALESCE(l.industry,'') AS industry,
       COALESCE(l.size,'') AS size, l.source, l.stage, l.assigned_to, l.mrr_potential,
       COALESCE(l.notes,'') AS notes, l.created_at, l.updated_at,
       CAST(COALESCE(DATE_FORMAT((SELECT t.next_follow_up FROM crm_touchpoints t
             WHERE t.lead_id = l.id AND t.business_id = l.business_id
             ORDER BY t.created_at DESC, t.id DESC LIMIT 1), '%Y-%m-%d'), '') AS CHAR) AS next_follow_up
FROM crm_leads l
WHERE l.id = ? AND l.business_id = ?;

-- name: CreateLead :execresult
INSERT INTO crm_leads (owner_id, business_id, company_name, industry, size, source, stage, assigned_to, mrr_potential, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateLead :exec
UPDATE crm_leads
SET company_name = ?, industry = ?, size = ?, source = ?, stage = ?,
    assigned_to = ?, mrr_potential = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteLead :execresult
DELETE FROM crm_leads WHERE id = ? AND business_id = ?;

-- name: GetPipeline :many
SELECT stage, COUNT(*) AS cnt, COALESCE(SUM(mrr_potential), 0) AS total_mrr
FROM crm_leads
WHERE business_id = ?
  AND created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 12 MONTH)
GROUP BY stage;

-- name: ListContacts :many
SELECT id, owner_id, business_id, lead_id, customer_id, name, COALESCE(role,'') AS role, COALESCE(email,'') AS email,
       COALESCE(phone,'') AS phone, COALESCE(linkedin,'') AS linkedin, is_primary, COALESCE(notes,'') AS notes,
       created_at, updated_at
FROM crm_contacts
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountContacts :one
SELECT COUNT(*) FROM crm_contacts WHERE business_id = ?;

-- name: ListContactsByLead :many
SELECT id, owner_id, business_id, lead_id, customer_id, name, COALESCE(role,'') AS role, COALESCE(email,'') AS email,
       COALESCE(phone,'') AS phone, COALESCE(linkedin,'') AS linkedin, is_primary, COALESCE(notes,'') AS notes,
       created_at, updated_at
FROM crm_contacts
WHERE business_id = ? AND lead_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountContactsByLead :one
SELECT COUNT(*) FROM crm_contacts WHERE business_id = ? AND lead_id = ?;

-- name: GetContact :one
SELECT id, owner_id, business_id, lead_id, customer_id, name, COALESCE(role,'') AS role, COALESCE(email,'') AS email,
       COALESCE(phone,'') AS phone, COALESCE(linkedin,'') AS linkedin, is_primary, COALESCE(notes,'') AS notes,
       created_at, updated_at
FROM crm_contacts WHERE id = ? AND business_id = ?;

-- name: CreateContact :execresult
INSERT INTO crm_contacts (owner_id, business_id, lead_id, customer_id, name, role, email, phone, linkedin, is_primary, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: LinkCustomerToContact :exec
UPDATE crm_contacts
SET customer_id = ?
WHERE business_id = ? AND id = ?;

-- name: UpdateContact :exec
UPDATE crm_contacts
SET customer_id = ?, name = ?, role = ?, email = ?, phone = ?, linkedin = ?, is_primary = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: DeleteContact :execresult
DELETE FROM crm_contacts WHERE id = ? AND business_id = ?;

-- name: ListTouchpoints :many
SELECT id, owner_id, business_id, lead_id, contact_id, type, outcome,
       COALESCE(notes,'') AS notes, next_follow_up, created_by, created_at
FROM crm_touchpoints
WHERE business_id = ? AND lead_id = ?
ORDER BY created_at DESC
LIMIT 100;

-- name: GetTouchpoint :one
SELECT id, owner_id, business_id, lead_id, contact_id, type, outcome,
       COALESCE(notes,'') AS notes, next_follow_up, created_by, created_at
FROM crm_touchpoints WHERE id = ? AND business_id = ?;

-- name: CreateTouchpoint :execresult
INSERT INTO crm_touchpoints (owner_id, business_id, lead_id, contact_id, type, outcome, notes, next_follow_up, created_by)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteTouchpoint :execresult
DELETE FROM crm_touchpoints WHERE id = ? AND business_id = ?;

-- name: ListContracts :many
SELECT id, owner_id, business_id, lead_id, plan, mrr, seats,
       value, billing, start_date, end_date, customer_id,
       signed_at, COALESCE(signed_by_name,'') AS signed_by_name, closed_by, COALESCE(notes,'') AS notes,
       created_at, updated_at
FROM crm_contracts
WHERE business_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountContracts :one
SELECT COUNT(*) FROM crm_contracts WHERE business_id = ?;

-- name: GetContract :one
SELECT id, owner_id, business_id, lead_id, plan, mrr, seats,
       value, billing, start_date, end_date, customer_id,
       signed_at, COALESCE(signed_by_name,'') AS signed_by_name, closed_by, COALESCE(notes,'') AS notes,
       created_at, updated_at
FROM crm_contracts WHERE id = ? AND business_id = ?;

-- name: CreateContract :execresult
INSERT INTO crm_contracts (owner_id, business_id, lead_id, plan, mrr, seats,
                           value, billing, start_date, end_date, customer_id,
                           signed_at, signed_by_name, closed_by, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateContract :exec
UPDATE crm_contracts
SET plan = ?, mrr = ?, seats = ?, value = ?, billing = ?, start_date = ?, end_date = ?,
    customer_id = ?, signed_at = ?, signed_by_name = ?, closed_by = ?, notes = ?
WHERE id = ? AND business_id = ?;

-- name: GetLeadCustomerID :one
-- The CRM customer a lead's contact became (contacts are captured into the
-- customer directory), preferring the primary contact.
SELECT customer_id FROM crm_contacts
WHERE lead_id = ? AND business_id = ? AND customer_id IS NOT NULL
ORDER BY is_primary DESC, id
LIMIT 1;

-- name: ListSignedContracts :many
-- Signed-only variant backing crm.sales.v1's signed filter (#839): a caller
-- asking for billable contracts must not page through drafts.
SELECT id, owner_id, business_id, lead_id, plan, mrr, seats,
       value, billing, start_date, end_date, customer_id,
       signed_at, COALESCE(signed_by_name,'') AS signed_by_name, closed_by, COALESCE(notes,'') AS notes,
       created_at, updated_at
FROM crm_contracts
WHERE business_id = ? AND signed_at IS NOT NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: ListUnsignedContracts :many
SELECT id, owner_id, business_id, lead_id, plan, mrr, seats,
       value, billing, start_date, end_date, customer_id,
       signed_at, COALESCE(signed_by_name,'') AS signed_by_name, closed_by, COALESCE(notes,'') AS notes,
       created_at, updated_at
FROM crm_contracts
WHERE business_id = ? AND signed_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountSignedContracts :one
SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND signed_at IS NOT NULL;

-- name: CountUnsignedContracts :one
SELECT COUNT(*) FROM crm_contracts WHERE business_id = ? AND signed_at IS NULL;
