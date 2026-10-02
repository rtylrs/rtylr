import React, { useId, useLayoutEffect, useRef, useState } from 'react';
import { isOverdue } from '../lib/dateUtils';
import { Calendar as CalendarIcon, X } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Calendar as CalendarGrid, todayISODate } from '@voxire/ui';
import { formatShortDate } from '../lib/utils';
import { toLocalIso } from '../lib/dateUtils';

interface DueDatePickerProps {
  value?: string;
  onChange: (date: string) => void;
  disabled?: boolean;
}

export const DueDatePicker: React.FC<DueDatePickerProps> = ({ value, onChange, disabled }) => {
  const { t, i18n } = useTranslation();
  const [open, setOpen] = useState(false);
  const popupId = useId();
  const direction = i18n.dir();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const popupRef = useRef<HTMLDivElement>(null);
  const close = () => {
    setOpen(false);
    triggerRef.current?.focus();
  };

  useLayoutEffect(() => {
    if (!open || disabled) return;
    const popup = popupRef.current;
    const trigger = triggerRef.current;
    if (!popup || !trigger) return;

    // The top layer escapes the task drawer's scrolling and clipping boundaries.
    popup.showPopover?.();
    const position = () => {
      const anchor = trigger.getBoundingClientRect();
      const width = popup.offsetWidth;
      const height = popup.offsetHeight;
      const gutter = 12;
      const alignedLeft = direction === 'rtl' ? anchor.right - width : anchor.left;
      const left = Math.max(gutter, Math.min(alignedLeft, window.innerWidth - width - gutter));
      const below = anchor.bottom + 6;
      const top =
        below + height <= window.innerHeight - gutter
          ? below
          : Math.max(
              gutter,
              Math.min(anchor.top - height - 6, window.innerHeight - height - gutter)
            );
      popup.style.left = `${left}px`;
      popup.style.top = `${top}px`;
    };
    position();
    popup.querySelector<HTMLButtonElement>('button')?.focus();
    const observer = new ResizeObserver(position);
    observer.observe(popup);
    const outside = (event: PointerEvent) => {
      if (
        event.target instanceof Node &&
        !popup.contains(event.target) &&
        !trigger.contains(event.target)
      ) {
        setOpen(false);
      }
    };
    const escape = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') return;
      event.preventDefault();
      event.stopPropagation();
      setOpen(false);
      trigger.focus();
    };
    const focusOutside = (event: FocusEvent) => {
      if (
        event.target instanceof Node &&
        !popup.contains(event.target) &&
        !trigger.contains(event.target)
      ) {
        setOpen(false);
      }
    };
    window.addEventListener('resize', position);
    window.addEventListener('scroll', position, true);
    window.addEventListener('keydown', escape, true);
    document.addEventListener('pointerdown', outside);
    document.addEventListener('focusin', focusOutside);
    return () => {
      observer.disconnect();
      window.removeEventListener('resize', position);
      window.removeEventListener('scroll', position, true);
      window.removeEventListener('keydown', escape, true);
      document.removeEventListener('pointerdown', outside);
      document.removeEventListener('focusin', focusOutside);
    };
  }, [open, disabled, direction]);
  const overdue = isOverdue(value);

  const quickOptions = [
    { id: 'today', label: t('flowCommon.datePicker.today'), days: 0 },
    { id: 'tomorrow', label: t('flowCommon.datePicker.tomorrow'), days: 1 },
    { id: 'in3Days', label: t('flowCommon.datePicker.in3Days'), days: 3 },
    { id: 'nextWeek', label: t('flowCommon.datePicker.nextWeek'), days: 7 },
    { id: 'in2Weeks', label: t('flowCommon.datePicker.in2Weeks'), days: 14 },
  ];

  return (
    <div className="relative">
      <div
        className="flex items-center gap-1.5 text-flow-meta"
        style={{ color: overdue ? 'var(--flow-danger-fg)' : 'var(--flow-text-muted)' }}
      >
        <button
          ref={triggerRef}
          type="button"
          aria-haspopup="dialog"
          aria-expanded={open && !disabled}
          aria-controls={open && !disabled ? popupId : undefined}
          onClick={() => !disabled && setOpen((o) => !o)}
          disabled={disabled}
          className="flow-due-date-trigger flex items-center gap-1.5 px-2 py-1 rounded-sm transition-colors hover:bg-[var(--flow-hover)] disabled:opacity-60 disabled:hover:bg-transparent"
        >
          <CalendarIcon size={14} />
          {value
            ? formatShortDate(value, t('flowCommon.dateLocale'))
            : t('flowCommon.datePicker.setDate')}
        </button>
        {value && !disabled && (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onChange('');
            }}
            aria-label={t('flowCommon.datePicker.clearDate')}
            className="flow-due-date-clear hover:text-[var(--flow-text)]"
          >
            <X size={10} />
          </button>
        )}
      </div>
      {open && !disabled && (
        <div
          ref={popupRef}
          id={popupId}
          popover="manual"
          role="dialog"
          aria-label={t('flow.taskDetail.dueDate')}
          dir={direction}
          className="flow-due-date-popover"
        >
          <div className="flow-due-date-shortcuts">
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
                    close();
                  }}
                  className="w-full text-start px-3 py-1.5 text-flow-meta hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {o.label}
                </button>
              );
            })}
          </div>
          <div style={{ borderTop: '1px solid var(--flow-border)', margin: '4px 0' }} />
          <CalendarGrid
            className="flow-due-date-calendar"
            value={value}
            onChange={(date) => {
              onChange(date);
              close();
            }}
            minDate={todayISODate()}
            locale={i18n.language}
            weekStartsOn={1}
            labels={{
              previousMonth: t('flowCommon.datePicker.previousMonth'),
              nextMonth: t('flowCommon.datePicker.nextMonth'),
            }}
          />
        </div>
      )}
    </div>
  );
};

export default DueDatePicker;
