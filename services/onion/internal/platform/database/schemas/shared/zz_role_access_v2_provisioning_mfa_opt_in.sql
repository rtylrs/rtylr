ALTER TABLE `user_provisioning_states`
  MODIFY COLUMN `mfa_required` tinyint(1) NOT NULL DEFAULT '0';

UPDATE `user_provisioning_states` AS provisioning
LEFT JOIN `users` AS account ON account.id = provisioning.user_id
SET provisioning.mfa_required = FALSE,
    provisioning.updated_at = NOW()
WHERE provisioning.mfa_required = TRUE
  AND provisioning.mfa_enrolled = FALSE
  AND COALESCE(account.two_factor_enabled, FALSE) = FALSE;
