import React, { useState } from 'react';
import { isOverdue } from '../lib/dateUtils';
import { Calendar as CalendarIcon, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Calendar as CalendarGrid, todayISODate } from '@voxire/ui';
import { formatDate } from '../lib/utils';
import { toLocalIso } from '../lib/dateUtils';

interface DueDatePickerProps {
  value?: string;
  onChange: (date: string) => void;
  disabled?: boolean;
}

export const DueDatePicker: React.FC<DueDatePickerProps> = ({ value, onChange, disabled }) => {
  const { t, i18n } = useTranslation();
  const [open, setOpen] = useState(false);
  // `isOverdue`, not `new Date(value) < new Date()`. A date-only string parses
  // as UTC midnight, so a date due **today** compared against the current
  // instant reads as past from 00:00 UTC, which is the whole working day at
  // UTC+3 (rtylr#1077). Three components had this; this was a fourth, missed
  // because the earlier sweep matched on the variable being named `dueDate`.
  const overdue = isOverdue(value);

  const quickOptions = [
    { id: 'today', label: t('common.datePicker.today'), days: 0 },
    { id: 'tomorrow', label: t('common.datePicker.tomorrow'), days: 1 },
    { id: 'in3Days', label: t('common.datePicker.in3Days'), days: 3 },
    { id: 'nextWeek', label: t('common.datePicker.nextWeek'), days: 7 },
    { id: 'in2Weeks', label: t('common.datePicker.in2Weeks'), days: 14 },
  ];

  return (
    <div className="relative">
      <div
        className="flex items-center gap-1.5 text-meta"
        style={{ color: overdue ? 'var(--flow-danger-fg)' : 'var(--flow-text-muted)' }}
      >
        <button
          type="button"
          onClick={() => !disabled && setOpen((o) => !o)}
          disabled={disabled}
          className="flex items-center gap-1.5 px-2 py-1 rounded-sm transition-colors hover:bg-[var(--flow-hover)] disabled:opacity-60 disabled:hover:bg-transparent"
        >
          <CalendarIcon size={12} />
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
            const iso = toLocalIso(d);
            return (
              <button
                type="button"
                key={o.id}
                onClick={() => {
                  onChange(iso);
                  setOpen(false);
                }}
                className="w-full text-start px-3 py-1.5 text-meta hover:bg-[var(--flow-hover)]"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {o.label}
              </button>
            );
          })}
          <div style={{ borderTop: '1px solid var(--flow-border)', margin: '4px 0' }} />
          <CalendarGrid
            value={value}
            onChange={(date) => {
              onChange(date);
              setOpen(false);
            }}
            minDate={todayISODate()}
            locale={i18n.language}
            weekStartsOn={1}
            labels={{
              previousMonth: t('common.datePicker.previousMonth'),
              nextMonth: t('common.datePicker.nextMonth'),
            }}
          />
        </div>
      )}
    </div>
  );
};

export default DueDatePicker;
