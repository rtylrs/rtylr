import React, { useId } from 'react';
import { useTranslation } from 'react-i18next';
import { BILLINGS } from '../data/model';

export const Field: React.FC<{
  label: string;
  children: (id: string) => React.ReactNode;
  hint?: string;
  error?: string;
  wide?: boolean;
}> = ({ label, children, hint, error, wide }) => {
  const id = useId();
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 4,
        gridColumn: wide ? '1 / -1' : undefined,
        minWidth: 0,
      }}
    >
      <label htmlFor={id} style={{ fontSize: 12, fontWeight: 600, color: 'var(--hub-text-muted)' }}>
        {label}
      </label>
      {children(id)}
      {error ? (
        <span role="alert" style={{ fontSize: 12, color: '#ef4444' }}>
          {error}
        </span>
      ) : (
        hint && <span style={{ fontSize: 12, color: 'var(--hub-text-dim)' }}>{hint}</span>
      )}
    </div>
  );
};

export const FieldGrid: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div
    style={{
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
      gap: 12,
    }}
  >
    {children}
  </div>
);

export const TextInput: React.FC<
  { id: string; value?: string; onChange: (value: string) => void } & Omit<
    React.InputHTMLAttributes<HTMLInputElement>,
    'value' | 'onChange' | 'id'
  >
> = ({ id, value, onChange, ...props }) => (
  <input
    id={id}
    className="sales-input"
    value={value ?? ''}
    onChange={(event) => onChange(event.target.value)}
    {...props}
  />
);

export const BillingSelect: React.FC<{
  id: string;
  value?: string;
  onChange: (value: string) => void;
}> = ({ id, value, onChange }) => {
  const { t } = useTranslation();
  return (
    <select
      id={id}
      className="sales-select"
      style={{ maxWidth: 'none' }}
      value={value || 'one_time'}
      onChange={(event) => onChange(event.target.value)}
    >
      {BILLINGS.map((billing) => (
        <option key={billing} value={billing}>
          {t(`crmSales.billing.${billing}`)}
        </option>
      ))}
    </select>
  );
};

/**
 * Reads a typed amount ("1,500.5") as the decimal string the API takes
 * ("1500.50"); "" stays empty and anything else is null.
 */
export function parseAmountInput(input: string): string | null {
  const cleaned = input.replace(/[\s,]/g, '');
  if (cleaned === '') return '';
  if (!/^\d+(\.\d{0,2})?$/.test(cleaned)) return null;
  return Number(cleaned).toFixed(2);
}

export function looksLikeEmail(value: string): boolean {
  return /^[^\s@,;]+@[^\s@,;]+$/.test(value.trim());
}
