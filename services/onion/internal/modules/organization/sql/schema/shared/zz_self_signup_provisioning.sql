ALTER TABLE `user_provisioning_states`
  MODIFY COLUMN `mode` enum('invitation','direct_credentials','self_signup') CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL;

-- Self-signup never wrote a provisioning row, and /users/me refuses an account
-- without one, so every account that signed up since provisioning landed was
-- locked out. Give those accounts the row signup now writes: active, with the
-- email still to verify.
INSERT INTO `user_provisioning_states` (user_id, business_id, mode, state, email_verified, password_change_required)
SELECT account.id, COALESCE(account.business_id, 0), 'self_signup', 'active', FALSE, FALSE
FROM `users` AS account
LEFT JOIN `user_provisioning_states` AS provisioning ON provisioning.user_id = account.id
WHERE provisioning.user_id IS NULL;
