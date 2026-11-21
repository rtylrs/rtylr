import React from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, ChevronRight, Inbox } from 'lucide-react';
import type { CalendarView } from './useCalendarView';
import { UNSCHEDULED_PANEL_ID } from './UnscheduledPanel';

interface CalendarControlsProps {
  view: CalendarView;
  onView: (view: CalendarView) => void;
  label: string;
  onStep: (direction: -1 | 1) => void;
  onToday: () => void;
  unscheduledCount: number;
  unscheduledOpen: boolean;
  onToggleUnscheduled: () => void;
}

const VIEWS: CalendarView[] = ['month', 'week'];

/**
 * Navigation for both calendars: previous, the period, next, Today, the Month
 * and Week switch, and the Unscheduled panel toggle.
 *
 * The chevrons are drawn for a left-to-right reading and flipped under `rtl:`,
 * so "previous" points backwards in Arabic too.
 */
export function CalendarControls({
  view,
  onView,
  label,
  onStep,
  onToday,
  unscheduledCount,
  unscheduledOpen,
  onToggleUnscheduled,
}: CalendarControlsProps): React.ReactElement {
  const { t } = useTranslation();
  const prevLabel =
    view === 'month' ? t('flow.calendar.prevMonthLabel') : t('flow.calendar.prevWeekLabel');
  const nextLabel =
    view === 'month' ? t('flow.calendar.nextMonthLabel') : t('flow.calendar.nextWeekLabel');
  const iconButton =
    'h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)] focus-visible:ring-1 focus-visible:ring-[var(--accent)]';

  return (
    <div className="flex flex-wrap items-center gap-2">
      <button
        type="button"
        onClick={() => onStep(-1)}
        aria-label={prevLabel}
        className={iconButton}
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <ChevronLeft size={14} className="rtl:-scale-x-100" />
      </button>
      <h2
        className="min-w-[128px] text-center text-flow-body font-semibold"
        style={{ color: 'var(--flow-text)' }}
      >
        {label}
      </h2>
      <button
        type="button"
        onClick={() => onStep(1)}
        aria-label={nextLabel}
        className={iconButton}
        style={{ color: 'var(--flow-text-muted)' }}
      >
        <ChevronRight size={14} className="rtl:-scale-x-100" />
      </button>
      <button
        type="button"
        onClick={onToday}
        className="rounded-sm px-2 py-1 text-flow-meta font-medium hover:bg-[var(--flow-hover)] focus-visible:ring-1 focus-visible:ring-[var(--accent)]"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {t('flow.calendar.today')}
      </button>
      <div
        role="group"
        aria-label={t('flow.calendar.viewLabel')}
        className="flex rounded-md p-0.5"
        style={{ border: '1px solid var(--flow-border)' }}
      >
        {VIEWS.map((option) => (
          <button
            key={option}
            type="button"
            aria-pressed={view === option}
            onClick={() => onView(option)}
            className="rounded-sm px-2 py-0.5 text-flow-meta font-medium focus-visible:ring-1 focus-visible:ring-[var(--accent)]"
            style={{
              background: view === option ? 'var(--flow-active)' : 'transparent',
              color: view === option ? 'var(--accent-on-muted)' : 'var(--flow-text-muted)',
            }}
          >
            {option === 'month' ? t('flow.calendar.month') : t('flow.calendar.week')}
          </button>
        ))}
      </div>
      <button
        type="button"
        aria-expanded={unscheduledOpen}
        aria-controls={unscheduledOpen ? UNSCHEDULED_PANEL_ID : undefined}
        onClick={onToggleUnscheduled}
        className="flex items-center gap-1.5 rounded-sm px-2 py-1 text-flow-meta font-medium hover:bg-[var(--flow-hover)] focus-visible:ring-1 focus-visible:ring-[var(--accent)]"
        style={{
          color: unscheduledOpen ? 'var(--flow-text)' : 'var(--flow-text-muted)',
          background: unscheduledOpen ? 'var(--flow-hover)' : 'transparent',
        }}
      >
        <Inbox size={13} />
        {t('flow.calendar.unscheduled.toggle', { count: unscheduledCount })}
      </button>
    </div>
  );
}
