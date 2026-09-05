import React from 'react';
import { Calendar } from 'lucide-react';
import { toIsoDate } from '../lib/dateUtils';

export interface DateRange {
  from: string;
  to: string;
}

interface DateRangePickerProps {
  value: DateRange;
  onChange: (range: DateRange) => void;
}

const presets: Array<{ label: string; days: number }> = [
  { label: 'Today', days: 0 },
  { label: '7 days', days: 7 },
  { label: '30 days', days: 30 },
  { label: '90 days', days: 90 },
];

const toISO = toIsoDate;

const DateRangePicker: React.FC<DateRangePickerProps> = ({ value, onChange }) => {
  const applyPreset = (days: number) => {
    const to = new Date();
    const from = new Date();
    from.setDate(from.getDate() - days);
    onChange({ from: toISO(from), to: toISO(to) });
  };

  return (
    <div
      className="inline-flex items-center gap-2 rounded-lg p-1"
      style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
    >
      <Calendar size={14} className="ml-2" style={{ color: 'var(--hub-text-dim)' }} />
      <input
        type="date"
        value={value.from}
        onChange={(e) => onChange({ ...value, from: e.target.value })}
        className="bg-transparent text-xs outline-none"
        style={{ color: 'var(--hub-text)' }}
      />
      <span className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
        to
      </span>
      <input
        type="date"
        value={value.to}
        onChange={(e) => onChange({ ...value, to: e.target.value })}
        className="bg-transparent text-xs outline-none"
        style={{ color: 'var(--hub-text)' }}
      />
      <div className="flex items-center gap-0.5 ml-2">
        {presets.map((p) => (
          <button
            type="button"
            key={p.label}
            onClick={() => applyPreset(p.days)}
            className="text-[10px] px-2 py-1 rounded"
            style={{ color: 'var(--hub-text-muted)' }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'var(--hub-active)';
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.background = 'transparent';
            }}
          >
            {p.label}
          </button>
        ))}
      </div>
    </div>
  );
};

export default DateRangePicker;
