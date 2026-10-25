-- HR setup completion belongs to the business. Rows used to be keyed by the
-- user who finished setup, so a manager's completion was invisible to the
-- owner and other admins. New rows carry the business ID in both owner_id
-- and business_id, the HR convention.
--
-- Legacy rows (business_id = 0) are first marked by negating owner_id so the
-- copies below cannot collide with them on the owner_id unique key. Each is
-- copied to every business its user owns or is an active member of, keeping
-- the earliest completion, then removed.

UPDATE hr_configs SET owner_id = -owner_id WHERE business_id = 0 AND owner_id > 0;

INSERT IGNORE INTO hr_configs (owner_id, business_id, setup_completed_at, created_at)
SELECT scope.business_id, scope.business_id, MIN(c.setup_completed_at), MIN(c.created_at)
FROM hr_configs c
JOIN (
    SELECT b.owner_id AS user_id, b.id AS business_id FROM businesses b
    UNION
    SELECT m.user_id, m.business_id FROM authorization_business_members m WHERE m.status = 'active'
) scope ON scope.user_id = -c.owner_id
WHERE c.business_id = 0 AND c.owner_id < 0
GROUP BY scope.business_id;

DELETE FROM hr_configs WHERE business_id = 0 AND owner_id < 0;

SET @apply := (SELECT COUNT(*) = 0 FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'hr_configs' AND index_name = 'uniq_hr_configs_business');
SET @stmt := IF(@apply, 'ALTER TABLE hr_configs ADD UNIQUE KEY uniq_hr_configs_business (business_id)', 'SELECT 1');
PREPARE hr_setup_stmt FROM @stmt;
EXECUTE hr_setup_stmt;
DEALLOCATE PREPARE hr_setup_stmt;
