ALTER TABLE `businesses`
  ADD COLUMN `mfa_enforced` tinyint(1) NOT NULL DEFAULT '0' AFTER `is_active`;
