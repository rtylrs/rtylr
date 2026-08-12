import React from 'react';

export type AccountType = 'asset' | 'liability' | 'equity' | 'revenue' | 'expense';

const STYLE: Record<AccountType, React.CSSProperties> = {
  asset: { background: 'var(--finance-info-soft)', color: 'var(--finance-info)' },
  liability: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
  equity: { background: 'var(--finance-purple-soft)', color: 'var(--finance-purple)' },
  revenue: { background: 'var(--finance-accent-soft)', color: 'var(--finance-accent)' },
  expense: { background: 'var(--finance-warning-soft)', color: 'var(--finance-warning)' },
};

const LABEL: Record<AccountType, string> = {
  asset: 'Asset',
  liability: 'Liability',
  equity: 'Equity',
  revenue: 'Revenue',
  expense: 'Expense',
};

const AccountTypeBadge: React.FC<{ type: AccountType }> = ({ type }) => (
  <span
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      padding: '3px 10px',
      borderRadius: 20,
      fontSize: 11,
      fontWeight: 600,
      ...STYLE[type],
    }}
  >
    {LABEL[type]}
  </span>
);

export default AccountTypeBadge;
