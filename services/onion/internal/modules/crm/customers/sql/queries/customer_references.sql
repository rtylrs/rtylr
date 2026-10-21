-- The CRM schema has no foreign keys, so deleting or merging a customer
-- updates every row that names them. Rows that only exist for the customer
-- are purged; sales records keep their history and lose or move the link.

-- name: DeleteCustomerCampaignSends :exec
DELETE FROM crm_campaign_sends WHERE business_id = ? AND customer_id = ?;

-- name: DeleteCustomerFeedback :exec
DELETE FROM crm_feedback WHERE business_id = ? AND customer_id = ?;

-- name: DeleteCustomerLoyaltyTransactions :exec
DELETE FROM crm_loyalty_transactions WHERE business_id = ? AND customer_id = ?;

-- name: DeleteCustomerTimeline :exec
DELETE FROM crm_customer_timeline WHERE business_id = ? AND customer_id = ?;

-- name: DetachCustomerFromContacts :exec
UPDATE crm_contacts SET customer_id = NULL WHERE business_id = ? AND customer_id = ?;

-- name: DetachCustomerFromLeads :exec
UPDATE crm_leads SET customer_id = NULL WHERE business_id = ? AND customer_id = ?;

-- name: DetachCustomerFromDeals :exec
UPDATE crm_deals SET customer_id = NULL WHERE business_id = ? AND customer_id = ?;

-- name: DetachCustomerFromActivities :exec
UPDATE crm_touchpoints SET customer_id = NULL WHERE business_id = ? AND customer_id = ?;

-- name: DetachCustomerFromContracts :exec
UPDATE crm_contracts SET customer_id = NULL WHERE business_id = ? AND customer_id = ?;

-- name: MoveCustomerCampaignSends :exec
UPDATE crm_campaign_sends SET customer_id = sqlc.arg('to_customer_id')
WHERE customer_id = sqlc.arg('from_customer_id') AND business_id = sqlc.arg('business_id');

-- name: MoveCustomerFeedback :exec
UPDATE crm_feedback SET customer_id = sqlc.arg('to_customer_id')
WHERE customer_id = sqlc.arg('from_customer_id') AND business_id = sqlc.arg('business_id');

-- name: MoveCustomerContacts :exec
UPDATE crm_contacts SET customer_id = sqlc.arg('to_customer_id')
WHERE customer_id = sqlc.arg('from_customer_id') AND business_id = sqlc.arg('business_id');

-- name: MoveCustomerLeads :exec
UPDATE crm_leads SET customer_id = sqlc.arg('to_customer_id')
WHERE customer_id = sqlc.arg('from_customer_id') AND business_id = sqlc.arg('business_id');

-- name: MoveCustomerDeals :exec
UPDATE crm_deals SET customer_id = sqlc.arg('to_customer_id')
WHERE customer_id = sqlc.arg('from_customer_id') AND business_id = sqlc.arg('business_id');

-- name: MoveCustomerActivities :exec
UPDATE crm_touchpoints SET customer_id = sqlc.arg('to_customer_id')
WHERE customer_id = sqlc.arg('from_customer_id') AND business_id = sqlc.arg('business_id');

-- name: MoveCustomerContracts :exec
UPDATE crm_contracts SET customer_id = sqlc.arg('to_customer_id')
WHERE customer_id = sqlc.arg('from_customer_id') AND business_id = sqlc.arg('business_id');
