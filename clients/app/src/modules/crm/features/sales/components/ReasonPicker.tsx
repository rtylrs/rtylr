import React from 'react';
import { useTranslation } from 'react-i18next';

/**
 * Common reasons as one-tap choices plus free text. The stored reason is the
 * text, so reports can group the common ones.
 */
export const ReasonPicker: React.FC<{
  reasons: readonly string[];
  /** i18n key prefix the reason ids are labelled under. */
  labelPrefix: string;
  value: string;
  onChange: (value: string) => void;
  label: string;
}> = ({ reasons, labelPrefix, value, onChange, label }) => {
  const { t } = useTranslation();
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div className="sales-chip-row" role="group" aria-label={label}>
        {reasons.map((reason) => {
          const text = t(`${labelPrefix}.${reason}`);
          return (
            <button
              key={reason}
              type="button"
              className="sales-chip"
              aria-pressed={value === text}
              onClick={() => onChange(text)}
            >
              {text}
            </button>
          );
        })}
      </div>
      <input
        className="sales-input"
        aria-label={label}
        placeholder={t('crmSales.common.otherReason')}
        value={value}
        maxLength={255}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  );
};
