import React, { useEffect, useRef, useState } from 'react';
import { CalendarDays, X } from 'lucide-react';
import { cn } from '../lib/cn';
import { Calendar, type CalendarLabels } from './Calendar';

export interface DatePickerLabels extends CalendarLabels {
  placeholder?: string;
  clear?: string;
}

export interface DatePickerProps {
  id?: string;
  value?: string;
  onChange: (date: string) => void;
  minDate?: string;
  maxDate?: string;
  weekStartsOn?: 0 | 1;
  locale?: string;
  disabled?: boolean;
  readOnly?: boolean;
  error?: boolean;
  clearable?: boolean;
  labels?: DatePickerLabels;
  formatValue?: (date: string) => string;
  className?: string;
  wrapperClassName?: string;
  'aria-label'?: string;
  'aria-labelledby'?: string;
  'aria-describedby'?: string;
  'aria-invalid'?: boolean;
}

const DatePicker = ({
  id,
  value,
  onChange,
  minDate,
  maxDate,
  weekStartsOn,
  locale,
  disabled,
  readOnly,
  error,
  clearable,
  labels,
  formatValue,
  className,
  wrapperClassName,
  'aria-label': ariaLabel,
  'aria-labelledby': ariaLabelledBy,
  'aria-describedby': ariaDescribedBy,
  'aria-invalid': ariaInvalid,
}: DatePickerProps): React.JSX.Element => {
  const inert = disabled || readOnly;
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onPointerDown = (event: MouseEvent | TouchEvent): void => {
      const target = event.target;
      const insidePicker = target instanceof Node && containerRef.current?.contains(target);
      if (!insidePicker) setOpen(false);
    };
    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('mousedown', onPointerDown);
    document.addEventListener('touchstart', onPointerDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('mousedown', onPointerDown);
      document.removeEventListener('touchstart', onPointerDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [open]);

  const display = value
    ? formatValue
      ? formatValue(value)
      : new Intl.DateTimeFormat(locale, { dateStyle: 'medium' }).format(
          new Date(`${value}T00:00:00Z`)
        )
    : (labels?.placeholder ?? 'Select date');

  return (
    <div ref={containerRef} className={cn('relative inline-block', wrapperClassName)}>
      <button
        id={id}
        type="button"
        disabled={inert}
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-label={ariaLabel}
        aria-labelledby={ariaLabelledBy}
        aria-describedby={ariaDescribedBy}
        aria-invalid={ariaInvalid}
        data-value={value ?? ''}
        onClick={() => setOpen((o) => !o)}
        className={cn(
          'flex h-10 min-h-[44px] w-full items-center justify-between gap-2 rounded-lg border px-3.5 text-sm outline-none transition-colors',
          'border-black/10 bg-white/[0.06] text-inherit dark:border-white/10',
          'hover:border-accent/60 focus-visible:border-accent',
          'disabled:cursor-not-allowed disabled:opacity-50',
          error && 'border-destructive focus-visible:border-destructive',
          !value && 'text-muted/70',
          className
        )}
      >
        <span className="flex items-center gap-2 truncate">
          <CalendarDays size={15} className="shrink-0 text-muted" />
          {display}
        </span>
        {clearable && value && !inert && (
          <span
            role="button"
            tabIndex={0}
            aria-label={labels?.clear ?? 'Clear date'}
            onClick={(e) => {
              e.stopPropagation();
              onChange('');
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                e.stopPropagation();
                onChange('');
              }
            }}
            className="shrink-0 rounded p-0.5 text-muted hover:text-inherit"
          >
            <X size={13} />
          </span>
        )}
      </button>
      {open && !inert && (
        <div
          role="dialog"
          aria-modal="false"
          className="absolute start-0 top-full z-50 mt-1 rounded-xl border border-black/10 bg-white shadow-dropdown dark:border-white/10 dark:bg-admin"
        >
          <Calendar
            value={value}
            onChange={(date) => {
              onChange(date);
              setOpen(false);
            }}
            minDate={minDate}
            maxDate={maxDate}
            weekStartsOn={weekStartsOn}
            locale={locale}
            labels={labels}
          />
        </div>
      )}
    </div>
  );
};

export { DatePicker };
