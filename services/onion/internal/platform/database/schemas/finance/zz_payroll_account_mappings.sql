-- Backfill mappings for legacy payroll accounts that this application created.
-- They remain ordinary tenant accounts and can be renamed or replaced; the
-- journal no longer creates any account implicitly.
CREATE TABLE IF NOT EXISTS finance_payroll_account_mappings (
    business_id BIGINT NOT NULL PRIMARY KEY,
    expense_account_id BIGINT NOT NULL,
    payable_account_id BIGINT NOT NULL,
    deductions_payable_account_id BIGINT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_finance_payroll_mapping_expense (expense_account_id),
    INDEX idx_finance_payroll_mapping_payable (payable_account_id),
    INDEX idx_finance_payroll_mapping_deductions (deductions_payable_account_id)
);

INSERT INTO finance_payroll_account_mappings (
    business_id, expense_account_id, payable_account_id, deductions_payable_account_id
)
SELECT expense.business_id,
       expense.id,
       payable.id,
       deductions.id
FROM finance_accounts expense
JOIN finance_accounts payable
  ON payable.business_id = expense.business_id
 AND payable.name = 'Payroll Payable'
 AND payable.type = 'liability'
 AND payable.is_active = TRUE
LEFT JOIN finance_accounts deductions
  ON deductions.business_id = expense.business_id
 AND deductions.name = 'Payroll Deductions Payable'
 AND deductions.type = 'liability'
 AND deductions.is_active = TRUE
WHERE expense.name = 'Payroll Expense'
  AND expense.type = 'expense'
  AND expense.is_active = TRUE
ON DUPLICATE KEY UPDATE
    expense_account_id = VALUES(expense_account_id),
    payable_account_id = VALUES(payable_account_id),
    deductions_payable_account_id = VALUES(deductions_payable_account_id),
    updated_at = CURRENT_TIMESTAMP;
