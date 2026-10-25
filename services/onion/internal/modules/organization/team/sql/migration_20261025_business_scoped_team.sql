-- Groups, group members and invitations belong to a business; their owner_id
-- copies of the business owner go nullable so this build can stop writing them,
-- and a later release drops them. Older invitations without a business take
-- the inviter's primary business, so every invitation names one.

DROP TEMPORARY TABLE IF EXISTS team_owner_primary_business;
CREATE TEMPORARY TABLE team_owner_primary_business AS
SELECT b.owner_id,
       COALESCE(
           MIN(CASE WHEN COALESCE(b.is_active, TRUE) = TRUE AND b.id = u.business_id THEN b.id END),
           MIN(CASE WHEN COALESCE(b.is_active, TRUE) = TRUE THEN b.id END),
           MIN(b.id)
       ) AS business_id
FROM businesses b
LEFT JOIN users u ON u.id = b.owner_id
GROUP BY b.owner_id;

SET @legacy_invites := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'invite_tokens' AND column_name = 'owner_id'
);

SET @stmt := IF(@legacy_invites,
    'UPDATE invite_tokens invite
     JOIN team_owner_primary_business p ON p.owner_id = invite.owner_id
     SET invite.business_id = p.business_id
     WHERE invite.business_id IS NULL',
    'SELECT 1');
PREPARE team_stmt FROM @stmt;
EXECUTE team_stmt;
DEALLOCATE PREPARE team_stmt;

DELETE FROM invite_tokens WHERE business_id IS NULL;

SET @nullable := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'invite_tokens'
      AND column_name = 'business_id' AND is_nullable = 'YES'
);
SET @stmt := IF(@nullable, 'ALTER TABLE invite_tokens MODIFY business_id bigint NOT NULL', 'SELECT 1');
PREPARE team_stmt FROM @stmt;
EXECUTE team_stmt;
DEALLOCATE PREPARE team_stmt;

SET @stmt := IF(@legacy_invites, 'ALTER TABLE invite_tokens MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE team_stmt FROM @stmt;
EXECUTE team_stmt;
DEALLOCATE PREPARE team_stmt;

SET @present := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'user_groups' AND column_name = 'owner_id'
);
SET @stmt := IF(@present, 'ALTER TABLE user_groups MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE team_stmt FROM @stmt;
EXECUTE team_stmt;
DEALLOCATE PREPARE team_stmt;

SET @present := (
    SELECT COUNT(*) > 0 FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'user_group_members' AND column_name = 'owner_id'
);
SET @stmt := IF(@present, 'ALTER TABLE user_group_members MODIFY owner_id bigint DEFAULT NULL', 'SELECT 1');
PREPARE team_stmt FROM @stmt;
EXECUTE team_stmt;
DEALLOCATE PREPARE team_stmt;

DROP TEMPORARY TABLE IF EXISTS team_owner_primary_business;
