import { useMemo, useState } from 'react';
import { monthCells, type MonthCell } from '../../shared/lib/monthGrid';
import { weekCells } from '../../shared/lib/calendarSpans';

export type CalendarView = 'month' | 'week';

/**
 * The month or week a calendar is showing, shared by Plan > Calendar and the
 * project calendar so the two step, label and range the same way.
 *
 * `from` and `to` are the first and last cells, padding included, because the
 * padding days are drawn and anything on them (a cycle, someone's leave) is
 * drawn with them.
 */
export function useCalendarView(language: string): {
  view: CalendarView;
  setView: (view: CalendarView) => void;
  cells: MonthCell[];
  from: string;
  to: string;
  label: string;
  step: (direction: -1 | 1) => void;
  goToday: () => void;
} {
  const [view, setView] = useState<CalendarView>('month');
  const [cursor, setCursor] = useState(() => new Date());

  const { cells, label } = useMemo(() => {
    if (view === 'month') {
      return {
        cells: monthCells(cursor),
        label: cursor.toLocaleString(language, { month: 'long', year: 'numeric' }),
      };
    }
    const week = weekCells(cursor);
    const first = week[0].date;
    const last = week[6].date;
    const dayMonth = new Intl.DateTimeFormat(language, { month: 'short', day: 'numeric' });
    const full = new Intl.DateTimeFormat(language, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
    return { cells: week, label: `${dayMonth.format(first)} – ${full.format(last)}` };
  }, [cursor, view, language]);

  const step = (direction: -1 | 1): void => {
    setCursor((current) =>
      view === 'month'
        ? new Date(current.getFullYear(), current.getMonth() + direction, 1)
        : new Date(current.getFullYear(), current.getMonth(), current.getDate() + 7 * direction)
    );
  };

  return {
    view,
    setView,
    cells,
    from: cells[0].iso,
    to: cells[cells.length - 1].iso,
    label,
    step,
    goToday: () => setCursor(new Date()),
  };
}
