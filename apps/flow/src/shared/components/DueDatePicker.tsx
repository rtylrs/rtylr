import React, { useState } from 'react';
import { Calendar, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { formatDate } from '../lib/utils';

interface DueDatePickerProps {
  value?: string;
  onChange: (date: string) => void;
  disabled?: boolean;
}

export const DueDatePicker: React.FC<DueDatePickerProps> = ({ value, onChange, disabled }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const overdue = value && new Date(value) < new Date();

  const quickOptions = [
    { label: t('common.datePicker.today'), days: 0 },
    { label: t('common.datePicker.tomorrow'), days: 1 },
    { label: t('common.datePicker.in3Days'), days: 3 },
    { label: t('common.datePicker.nextWeek'), days: 7 },
    { label: t('common.datePicker.in2Weeks'), days: 14 },
  ];

  return (
    <div className="relative">
      <div
        className="flex items-center gap-1.5 text-[11px]"
        style={{ color: overdue ? '#ef4444' : 'var(--flow-text-muted)' }}
      >
        <button
          type="button"
          onClick={() => !disabled && setOpen((o) => !o)}
          disabled={disabled}
          className="flex items-center gap-1.5 px-2 py-1 rounded transition-colors hover:bg-[var(--flow-hover)] disabled:opacity-60 disabled:hover:bg-transparent"
        >
          <Calendar size={12} />
          {value ? formatDate(value) : t('common.datePicker.setDate')}
        </button>
        {value && !disabled && (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onChange('');
            }}
            aria-label={t('common.datePicker.clearDate')}
            className="hover:text-[var(--flow-text)]"
          >
            <X size={10} />
          </button>
        )}
      </div>
      {open && !disabled && (
        <div
          className="absolute top-full start-0 mt-1 rounded-lg z-50 py-1 min-w-[160px]"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
            boxShadow: 'var(--flow-shadow)',
          }}
        >
          {quickOptions.map((o) => {
            const d = new Date();
            d.setDate(d.getDate() + o.days);
            const iso = d.toISOString().split('T')[0];
            return (
              <button
                key={o.label}
                onClick={() => {
                  onChange(iso);
                  setOpen(false);
                }}
                className="w-full text-start px-3 py-1.5 text-[12px] hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {o.label}
              </button>
            );
          })}
          <div style={{ borderTop: '1px solid var(--flow-border)', margin: '4px 0' }} />
          <div className="px-3 py-1.5">
            <input
              type="date"
              value={value || ''}
              min={new Date().toISOString().split('T')[0]}
              onChange={(e) => {
                onChange(e.target.value);
                setOpen(false);
              }}
              className="bg-transparent text-[12px] w-full outline-none"
              style={{ color: 'var(--flow-text)' }}
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default DueDatePicker;
