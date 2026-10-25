-- Billing belongs to the business, not to whoever owns it, so a business keeps
-- its plan, seat ledger and plan grants when ownership moves. Existing rows move
-- to the owner's primary business, and each of the owner's other active
-- businesses gets a copy of the plan and live grants, except a plan Paddle
-- bills: nothing would pay for or cancel the copies, so those businesses start
-- on the default plan instead. owner_id and the grants' user_id stay, nullable,
-- until a later release drops them: the previous build still reads them while
-- this one rolls out. Fresh databases already have the final shape, so every
-- step below is a no-op there.

SET @legacy_billing := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'subscriptions' AND column_name = 'owner_id'
);

DROP TEMPORARY TABLE IF EXISTS billing_owner_primary_business;
CREATE TEMPORARY TABLE billing_owner_primary_business AS
SELECT b.owner_id,
       COALESCE(
           MIN(CASE WHEN COALESCE(b.is_active, TRUE) = TRUE AND b.id = u.business_id THEN b.id END),
           MIN(CASE WHEN COALESCE(b.is_active, TRUE) = TRUE THEN b.id END),
           MIN(b.id)
       ) AS business_id
FROM businesses b
LEFT JOIN users u ON u.id = b.owner_id
GROUP BY b.owner_id;

-- subscriptions ---------------------------------------------------------------

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'subscriptions' AND column_name = 'business_id'
);
SET @stmt := IF(@missing, 'ALTER TABLE subscriptions ADD COLUMN business_id bigint DEFAULT NULL AFTER id', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @stmt := IF(@legacy_billing, 'ALTER TABLE subscriptions MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @present := (
    SELECT COUNT(*) > 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'subscriptions' AND index_name = 'owner_id'
);
SET @stmt := IF(@present, 'ALTER TABLE subscriptions DROP INDEX owner_id', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @stmt := IF(@legacy_billing,
    'UPDATE subscriptions s
     JOIN billing_owner_primary_business p ON p.owner_id = s.owner_id
     SET s.business_id = p.business_id
     WHERE s.business_id IS NULL
       AND p.business_id NOT IN (
           SELECT taken.business_id
           FROM (SELECT DISTINCT business_id FROM subscriptions WHERE business_id IS NOT NULL) taken
       )',
    'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @stmt := IF(@legacy_billing,
    'INSERT INTO subscriptions (business_id, owner_id, plan_id, status, seat_count, price_per_seat_cents,
                                annual_price_per_seat_cents, trial_ends_at, next_billing_date, current_period_ends_at)
     SELECT b.id, s.owner_id, s.plan_id, s.status, s.seat_count, s.price_per_seat_cents,
            s.annual_price_per_seat_cents, s.trial_ends_at, s.next_billing_date, s.current_period_ends_at
     FROM subscriptions s
     JOIN billing_owner_primary_business p ON p.owner_id = s.owner_id AND p.business_id = s.business_id
     JOIN businesses b ON b.owner_id = s.owner_id AND COALESCE(b.is_active, TRUE) = TRUE AND b.id <> s.business_id
     WHERE COALESCE(s.paddle_subscription_id, '''') = ''''
       AND NOT EXISTS (SELECT 1 FROM subscriptions taken WHERE taken.business_id = b.id)',
    'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'subscriptions' AND index_name = 'uk_subscriptions_business'
);
SET @stmt := IF(@missing, 'ALTER TABLE subscriptions ADD UNIQUE KEY uk_subscriptions_business (business_id)', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

-- seat_periods ----------------------------------------------------------------

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'seat_periods' AND column_name = 'business_id'
);
SET @stmt := IF(@missing, 'ALTER TABLE seat_periods ADD COLUMN business_id bigint DEFAULT NULL AFTER subscription_id', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @present := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'seat_periods' AND column_name = 'owner_id'
);
SET @stmt := IF(@present, 'ALTER TABLE seat_periods MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

UPDATE seat_periods sp
JOIN subscriptions s ON s.id = sp.subscription_id
SET sp.business_id = s.business_id
WHERE sp.business_id IS NULL AND s.business_id IS NOT NULL;

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'seat_periods' AND index_name = 'idx_sp_business_month'
);
SET @stmt := IF(@missing, 'ALTER TABLE seat_periods ADD KEY idx_sp_business_month (business_id, billing_month)', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

-- manual_access_grants --------------------------------------------------------

SET @legacy_grants := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'manual_access_grants' AND column_name = 'user_id'
);

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'manual_access_grants' AND column_name = 'business_id'
);
SET @stmt := IF(@missing, 'ALTER TABLE manual_access_grants ADD COLUMN business_id bigint DEFAULT NULL AFTER id', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @stmt := IF(@legacy_grants, 'ALTER TABLE manual_access_grants MODIFY user_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @stmt := IF(@legacy_grants,
    'UPDATE manual_access_grants g
     JOIN billing_owner_primary_business p ON p.owner_id = g.user_id
     SET g.business_id = p.business_id
     WHERE g.business_id IS NULL',
    'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @stmt := IF(@legacy_grants,
    'INSERT INTO manual_access_grants (business_id, user_id, plan_id, granted_by, note, expires_at, created_at)
     SELECT b.id, g.user_id, g.plan_id, g.granted_by, g.note, g.expires_at, g.created_at
     FROM manual_access_grants g
     JOIN billing_owner_primary_business p ON p.owner_id = g.user_id AND p.business_id = g.business_id
     JOIN businesses b ON b.owner_id = g.user_id AND COALESCE(b.is_active, TRUE) = TRUE AND b.id <> g.business_id
     WHERE g.revoked_at IS NULL
       AND NOT EXISTS (
           SELECT 1 FROM manual_access_grants copied
           WHERE copied.business_id = b.id AND copied.plan_id = g.plan_id AND copied.created_at <=> g.created_at
       )',
    'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

SET @missing := (
    SELECT COUNT(*) = 0 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = 'manual_access_grants' AND index_name = 'idx_mag_business_plan'
);
SET @stmt := IF(@missing, 'ALTER TABLE manual_access_grants ADD KEY idx_mag_business_plan (business_id, plan_id)', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

-- manual_addon_grants already carries business_id; only its owner column goes.
SET @present := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'manual_addon_grants' AND column_name = 'user_id'
);
SET @stmt := IF(@present, 'ALTER TABLE manual_addon_grants MODIFY user_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE billing_stmt FROM @stmt;
EXECUTE billing_stmt;
DEALLOCATE PREPARE billing_stmt;

DROP TEMPORARY TABLE IF EXISTS billing_owner_primary_business;
