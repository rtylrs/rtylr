import React from 'react';
import { useTranslation } from 'react-i18next';

interface Props {
  from: string;
  to: string;
  onChange: (from: string, to: string) => void;
}

const DateRangeFilter: React.FC<Props> = ({ from, to, onChange }) => {
  const { t } = useTranslation();

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      <input
        type="date"
        value={from}
        onChange={(e) => onChange(e.target.value, to)}
        aria-label={t('common.dateRangeFrom')}
        style={{
          padding: '6px 8px',
          borderRadius: 8,
          border: '1px solid var(--hr-border)',
          background: 'var(--hr-input-bg)',
          color: 'var(--hr-text)',
          fontSize: 16,
        }}
      />
      <span style={{ color: 'var(--hr-text-muted)', fontSize: 11 }}>{t('common.dateRangeTo')}</span>
      <input
        type="date"
        value={to}
        onChange={(e) => onChange(from, e.target.value)}
        aria-label={t('common.dateRangeTo')}
        style={{
          padding: '6px 8px',
          borderRadius: 8,
          border: '1px solid var(--hr-border)',
          background: 'var(--hr-input-bg)',
          color: 'var(--hr-text)',
          fontSize: 16,
        }}
      />
    </div>
  );
};

export default DateRangeFilter;
