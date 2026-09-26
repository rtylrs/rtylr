

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
