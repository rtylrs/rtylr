-- The bank account list reads every column below, and production answered it
-- with a 500. Two of them (last_four, is_primary) were added by an ALTER file
-- that existed for two days before the schema consolidation folded it into the
-- CREATE TABLE in finance/schema.sql. A database whose migrations were halted
-- in that window recorded schema.sql as applied without ever getting them, and
-- nothing would add them since. Add any column the table is missing; on a
-- database that already has them every statement is a no-op.

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'business_id'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN business_id bigint NOT NULL DEFAULT ''0'' AFTER owner_id'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'institution'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN institution varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER name'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'account_type'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN account_type varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER institution'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'last_four'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN last_four varchar(4) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER account_type'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'balance'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN balance decimal(15,2) DEFAULT ''0.00'' AFTER last_four'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'currency'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN currency varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT ''USD'' AFTER balance'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'is_primary'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN is_primary tinyint(1) NOT NULL DEFAULT ''0'' AFTER currency'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'is_connected'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN is_connected tinyint(1) DEFAULT ''0'' AFTER is_primary'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;

SET @bank_account_column_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_bank_accounts'
      AND column_name = 'last_synced_at'
);
SET @bank_account_column_sql := IF(
    @bank_account_column_exists,
    'SELECT 1',
    'ALTER TABLE finance_bank_accounts ADD COLUMN last_synced_at timestamp NULL DEFAULT NULL AFTER is_connected'
);
PREPARE bank_account_column FROM @bank_account_column_sql;
EXECUTE bank_account_column;
DEALLOCATE PREPARE bank_account_column;
