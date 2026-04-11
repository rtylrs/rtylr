import React from 'react';

export type AccountType = 'asset' | 'liability' | 'equity' | 'revenue' | 'expense';

const STYLE: Record<AccountType, React.CSSProperties> = {
  asset: { background: 'rgba(59,130,246,0.12)', color: '#3b82f6' },
  liability: { background: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  equity: { background: 'rgba(139,92,246,0.12)', color: '#8b5cf6' },
  revenue: { background: 'rgba(16,185,129,0.12)', color: '#10b981' },
  expense: { background: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
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
