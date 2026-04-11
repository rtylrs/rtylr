import React from 'react';

export type PaymentStatus = 'paid' | 'partial' | 'unpaid' | 'overdue' | 'refunded';

const STYLE: Record<PaymentStatus, React.CSSProperties> = {
  paid: { background: 'rgba(16,185,129,0.12)', color: '#10b981' },
  partial: { background: 'rgba(245,158,11,0.12)', color: '#f59e0b' },
  unpaid: { background: 'rgba(113,113,122,0.15)', color: '#a1a1aa' },
  overdue: { background: 'rgba(239,68,68,0.12)', color: '#ef4444' },
  refunded: { background: 'rgba(139,92,246,0.12)', color: '#8b5cf6' },
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
