-- Expense categories were created without a uniqueness check, so a form that
-- submitted twice left a business with several categories of the same name
-- that nobody could tell apart. Keep the oldest category of each name, move
-- every expense onto it, drop the copies, and let the database refuse new
-- copies from now on. The column collation ignores case and trailing spaces,
-- so "Rent" and "rent " count as one name, the same rule the API applies.

-- The keeper subquery aggregates, so MySQL materialises it before the DELETE
-- touches the table; no temporary table is needed, which keeps the file safe
-- under GTID consistency on the hosted database.
UPDATE finance_expenses e
JOIN finance_expense_categories copy ON copy.id = e.category_id
JOIN (
    SELECT business_id, name, MIN(id) AS keep_id
    FROM finance_expense_categories
    GROUP BY business_id, name
    HAVING COUNT(*) > 1
) keeper ON keeper.business_id = copy.business_id AND keeper.name = copy.name
SET e.category_id = keeper.keep_id
WHERE copy.id <> keeper.keep_id;

DELETE copy FROM finance_expense_categories copy
JOIN (
    SELECT business_id, name, MIN(id) AS keep_id
    FROM finance_expense_categories
    GROUP BY business_id, name
    HAVING COUNT(*) > 1
) keeper ON keeper.business_id = copy.business_id AND keeper.name = copy.name
WHERE copy.id <> keeper.keep_id;

SET @expense_category_unique_exists := (
    SELECT COUNT(*) > 0
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'finance_expense_categories'
      AND index_name = 'uq_finance_expense_categories_business_name'
);
SET @expense_category_unique_sql := IF(
    @expense_category_unique_exists,
    'SELECT 1',
    'ALTER TABLE finance_expense_categories ADD UNIQUE KEY uq_finance_expense_categories_business_name (business_id, name)'
);
PREPARE expense_category_unique FROM @expense_category_unique_sql;
EXECUTE expense_category_unique;
DEALLOCATE PREPARE expense_category_unique;
