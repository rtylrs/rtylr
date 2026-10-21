-- name: ListCampaigns :many
SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,
       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,
       opened_count, clicked_count, created_at, updated_at
FROM crm_campaigns WHERE business_id = ? ORDER BY created_at DESC LIMIT 100;

-- name: ListCampaignsByStatus :many
SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,
       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,
       opened_count, clicked_count, created_at, updated_at
FROM crm_campaigns WHERE business_id = ? AND status = ? ORDER BY created_at DESC LIMIT 100;

-- name: GetCampaignByID :one
SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,
       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,
       opened_count, clicked_count, created_at, updated_at
FROM crm_campaigns WHERE id = ? AND business_id = ?;

-- name: CreateCampaign :execresult
INSERT INTO crm_campaigns (owner_id, business_id, name, description, channel, segment_id, subject, content, scheduled_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateCampaign :execresult
UPDATE crm_campaigns
SET name = ?, description = ?, channel = ?, segment_id = ?, subject = ?,
    content = ?, scheduled_at = ?
WHERE id = ? AND business_id = ?;

-- name: UpdateCampaignStatus :execresult
UPDATE crm_campaigns SET status = ? WHERE id = ? AND business_id = ?;

-- name: ListCampaignsPaginated :many
SELECT id, owner_id, business_id, name, description, channel, status, segment_id, subject,
       content, scheduled_at, sent_at, audience_count, sent_count, delivered_count,
       opened_count, clicked_count, created_at, updated_at
FROM crm_campaigns WHERE business_id = ? ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountCampaigns :one
SELECT COUNT(*) FROM crm_campaigns WHERE business_id = ?;
