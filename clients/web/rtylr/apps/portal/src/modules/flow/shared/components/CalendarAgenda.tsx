import React from 'react';
import { useTranslation } from 'react-i18next';
import { onTint } from '../lib/colorUtils';
import type { MonthCell } from '../lib/monthGrid';

export interface AgendaEntry<T> {
  cell: MonthCell;
  tasks: T[];
}

interface CalendarAgendaProps<T> {
  entries: AgendaEntry<T>[];
  /** `YYYY-MM-DD` for the reader's today, so the heading can say so. */
  todayIso: string;
  /** Stable key and click target. */
  taskId: (task: T) => string;
  taskTitle: (task: T) => string;
  /**
   * The colour this calendar uses to say what a task is: the project on the
   * global calendar, the status on a project's own. Returning undefined leaves
   * the row untinted rather than painting it with `undefined`.
   */
  taskTone: (task: T) => string | undefined;
  onOpen: (task: T) => void;
  /** Rendered before the title, e.g. the priority glyph. */
  leading?: (task: T) => React.ReactNode;
}

/**
 * The month as a list, for viewports too narrow for a seven-column grid.
 *
 * At 390px a month grid gives each day about 55px, which truncates a task title
 * to two or three characters and leaves the pills too small to hit reliably
 * (rtylr#1215, S-10). This is the same information as a list: a heading per day
 * that has tasks, then that day's tasks as 40px rows.
 *
 * Only days with tasks appear. Thirty headings on a phone, most of them empty,
 * would be the grid's problem in a taller shape.
 *
 * Row geometry follows S-04, the title on one line with an ellipsis, but at 44px
 * rather than its 40px. S-04 is a desktop list where the row is not the tap
 * target; here the whole row is the button, so it takes the 44px minimum. The
 * mock is the authority on the shape, not on a number that would break the
 * thing the shape is for.
 */
export function CalendarAgenda<T>({
  entries,
  todayIso,
  taskId,
  taskTitle,
  taskTone,
  onOpen,
  leading,
}: CalendarAgendaProps<T>): React.JSX.Element {
  const { t, i18n } = useTranslation();

  if (entries.length === 0) {
    return (
      <p
        className="px-4 py-8 text-center text-flow-meta"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {t('flow.calendar.agendaEmpty')}
      </p>
    );
  }

  const heading = new Intl.DateTimeFormat(i18n.language, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });

  return (
    <ul className="flex flex-col">
      {entries.map(({ cell, tasks }) => {
        const isToday = cell.iso === todayIso;
        return (
          <li key={cell.iso}>
            <h3
              className="sticky top-0 z-10 flex items-baseline gap-2 px-4 py-2 text-sm font-semibold"
              style={{
                background: 'var(--flow-bg)',
                borderBottom: '1px solid var(--flow-border)',
                color: isToday ? 'var(--accent)' : 'var(--flow-text)',
              }}
            >
              {heading.format(cell.date)}
              {isToday && (
                <span className="text-flow-meta font-medium" style={{ color: 'var(--accent)' }}>
                  {t('flow.calendar.today')}
                </span>
              )}
            </h3>
            <ul>
              {tasks.map((task) => {
                const tone = taskTone(task);
                return (
                  <li key={taskId(task)}>
                    <button
                      type="button"
                      onClick={() => onOpen(task)}
                      className="flow-focus flex h-11 w-full items-center gap-3 px-4 text-start"
                      style={{ borderBottom: '1px solid var(--flow-border)' }}
                    >
                      {/* The tint cue the grid's pills carry, kept so the list
                          says the same thing the grid did (rtylr#1214). */}
                      <span
                        aria-hidden="true"
                        className="h-2 w-2 shrink-0 rounded-full"
                        style={{ background: tone ?? 'var(--flow-border-control)' }}
                      />
                      {leading?.(task)}
                      {/* dir="auto" so a right-to-left title truncates at its own
                          end rather than clipping the start of the phrase. */}
                      <span
                        dir="auto"
                        className="min-w-0 flex-1 truncate text-sm"
                        style={
                          tone
                            ? { ['--tone' as string]: tone, color: onTint('var(--tone)') }
                            : { color: 'var(--flow-text)' }
                        }
                      >
                        {taskTitle(task)}
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          </li>
        );
      })}
    </ul>
  );
}

export default CalendarAgenda;
