import React from 'react';

export type PaymentStatus = 'paid' | 'partial' | 'unpaid' | 'overdue' | 'refunded';

const STYLE: Record<PaymentStatus, React.CSSProperties> = {
  paid: { background: 'var(--finance-success-soft)', color: 'var(--finance-success)' },
  partial: { background: 'var(--finance-warning-soft)', color: 'var(--finance-warning)' },
  unpaid: { background: 'var(--finance-neutral-soft)', color: 'var(--finance-text-dim)' },
  overdue: { background: 'var(--finance-danger-soft)', color: 'var(--finance-danger)' },
  refunded: { background: 'var(--finance-purple-soft)', color: 'var(--finance-purple)' },
};

const LABEL: Record<PaymentStatus, string> = {
  paid: 'Paid',
  partial: 'Partial',
  unpaid: 'Unpaid',
  overdue: 'Overdue',
  refunded: 'Refunded',
};

const PaymentBadge: React.FC<{ status: PaymentStatus; compact?: boolean }> = ({
  status,
  compact,
}) => (
  <span
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      padding: compact ? '2px 8px' : '3px 10px',
      borderRadius: 20,
      fontSize: compact ? 10 : 11,
      fontWeight: 600,
      ...STYLE[status],
    }}
  >
    {LABEL[status]}
  </span>
);

export default PaymentBadge;
