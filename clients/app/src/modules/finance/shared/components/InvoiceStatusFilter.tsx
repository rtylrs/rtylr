import React from 'react';
import { ACCENT } from '../lib/tokens';

export type InvoiceStatus = 'Draft' | 'Sent' | 'Paid' | 'Overdue';
export type InvoiceFilterValue = InvoiceStatus | 'All';

interface InvoiceStatusFilterProps {
  value: InvoiceFilterValue;
  onChange: (v: InvoiceFilterValue) => void;
  counts?: Partial<Record<InvoiceFilterValue, number>>;
}

const TABS: InvoiceFilterValue[] = ['All', 'Draft', 'Sent', 'Paid', 'Overdue'];

const InvoiceStatusFilter: React.FC<InvoiceStatusFilterProps> = ({ value, onChange, counts }) => (
  <div
    style={{
      display: 'flex',
      gap: 4,
      padding: 4,
      background: 'var(--hub-surface)',
      borderRadius: 10,
      border: '1px solid var(--hub-border)',
      width: 'fit-content',
    }}
  >
    {TABS.map((tab) => (
      <button
        type="button"
        key={tab}
        onClick={() => onChange(tab)}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          padding: '5px 14px',
          borderRadius: 7,
          border: 'none',
          fontSize: 12,
          fontWeight: 500,
          cursor: 'pointer',
          background: value === tab ? ACCENT : 'transparent',
          color: value === tab ? 'var(--finance-white)' : 'var(--hub-text-muted)',
          transition: 'all 0.15s',
        }}
      >
        {tab}
        {counts?.[tab] !== undefined && (
          <span
            style={{
              fontSize: 10,
              padding: '1px 6px',
              borderRadius: 10,
              background: value === tab ? 'var(--finance-white-overlay)' : 'var(--hub-hover)',
            }}
          >
            {counts[tab]}
          </span>
        )}
      </button>
    ))}
  </div>
);

export default InvoiceStatusFilter;
