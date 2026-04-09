import React from 'react';

interface Props {
  from: string;
  to: string;
  onChange: (from: string, to: string) => void;
}

const DateRangeFilter: React.FC<Props> = ({ from, to, onChange }) => (
  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
    <input
      type="date"
      value={from}
      onChange={(e) => onChange(e.target.value, to)}
      style={{
        padding: '6px 8px',
        borderRadius: 8,
        border: '1px solid var(--hr-border)',
        background: 'var(--hr-input-bg)',
        color: 'var(--hr-text)',
        fontSize: 12,
      }}
    />
    <span style={{ color: 'var(--hr-text-muted)', fontSize: 11 }}>to</span>
    <input
      type="date"
      value={to}
      onChange={(e) => onChange(from, e.target.value)}
      style={{
        padding: '6px 8px',
        borderRadius: 8,
        border: '1px solid var(--hr-border)',
        background: 'var(--hr-input-bg)',
        color: 'var(--hr-text)',
        fontSize: 12,
      }}
    />
  </div>
);

export default DateRangeFilter;
