-- Business owners created from the console (partials "create business owner")
-- never got a provisioning row, and /users/me refuses an account without one,
-- so each of them signed in and then hit "Couldn't load your account".
-- CreateBusinessOwner now writes the row. This gives the owners created before
-- that the same row: active, and email-verified, because an operator typed the
-- address and set the password.
--
-- Only owners with no row are touched; self-signup accounts were backfilled
-- (unverified) by zz_self_signup_provisioning.sql and keep their state.
INSERT INTO `user_provisioning_states` (user_id, business_id, mode, state, email_verified, password_change_required)
SELECT business.owner_id, MIN(business.id), 'direct_credentials', 'active', TRUE, FALSE
FROM `businesses` AS business
LEFT JOIN `user_provisioning_states` AS provisioning ON provisioning.user_id = business.owner_id
WHERE provisioning.user_id IS NULL
  AND business.owner_id IS NOT NULL
GROUP BY business.owner_id;
