import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Search, X } from 'lucide-react';
import { withAlpha, onTint } from '../../shared/lib/colorUtils';
import { isUnscheduled } from '../../shared/lib/calendarSpans';
import type { CalendarScheduling } from './useCalendarScheduling';
import type { GridTask } from './PlanningGrid';

export const UNSCHEDULED_PANEL_ID = 'flow-calendar-unscheduled';

/** A task still worth scheduling: no dates, not finished, not archived. */
export function needsScheduling(task: {
  dueDate?: string | null;
  startDate?: string | null;
  status?: string;
  archived?: boolean;
}): boolean {
  if (task.archived) return false;
  if (task.status === 'done' || task.status === 'cancelled') return false;
  return isUnscheduled(task);
}

interface UnscheduledPanelProps<T extends GridTask> {
  /** The calendar's tasks; the panel keeps the ones with no dates. */
  tasks: readonly T[];
  tone: (task: T) => string | undefined;
  canEdit: (task: T) => boolean;
  onOpen: (task: T) => void;
  onClose: () => void;
  scheduling: Pick<CalendarScheduling<T>, 'begin' | 'end'>;
  /** `side` beside the grid; `stacked` full width above the narrow list. */
  layout: 'side' | 'stacked';
}

/**
 * Open tasks with neither a start nor a due date, to drag onto a day.
 *
 * Dropping one on a day sets its due date through the ordinary task update. A
 * task the user cannot edit is listed but not draggable, so a viewer can still
 * find and open it. On the narrow layout there is no grid to drop on, so the
 * list is for finding and opening only.
 */
export function UnscheduledPanel<T extends GridTask & { status?: string; archived?: boolean }>({
  tasks,
  tone,
  canEdit,
  onOpen,
  onClose,
  scheduling,
  layout,
}: UnscheduledPanelProps<T>): React.ReactElement {
  const { t } = useTranslation();
  const [query, setQuery] = useState('');

  const unscheduled = useMemo(() => tasks.filter(needsScheduling), [tasks]);
  const shown = useMemo(() => {
    const needle = query.trim().toLocaleLowerCase();
    if (!needle) return unscheduled;
    return unscheduled.filter((task) => task.title.toLocaleLowerCase().includes(needle));
  }, [unscheduled, query]);

  const draggable = layout === 'side';
  const anyEditable = draggable && unscheduled.some(canEdit);

  return (
    <aside
      id={UNSCHEDULED_PANEL_ID}
      aria-label={t('flow.calendar.unscheduled.title')}
      className={
        layout === 'side'
          ? 'flex w-64 shrink-0 flex-col min-h-0 border-s'
          : 'flex max-h-72 flex-col border-b'
      }
      style={{ borderColor: 'var(--flow-border)', background: 'var(--flow-bg)' }}
    >
      <div className="flex items-center justify-between gap-2 px-3 pt-3">
        <h2 className="text-flow-body font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('flow.calendar.unscheduled.title')}
          <span
            className="ms-1.5 text-flow-meta font-normal"
            style={{ color: 'var(--flow-text-dim)' }}
          >
            {unscheduled.length}
          </span>
        </h2>
        <button
          type="button"
          onClick={onClose}
          aria-label={t('flow.calendar.unscheduled.hide')}
          className="flex h-7 w-7 items-center justify-center rounded-sm hover:bg-[var(--flow-hover)] focus-visible:ring-1 focus-visible:ring-[var(--accent)]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <X size={12} />
        </button>
      </div>
      <p className="px-3 pt-1 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
        {anyEditable
          ? t('flow.calendar.unscheduled.hint')
          : t('flow.calendar.unscheduled.readOnly')}
      </p>
      <label className="relative mx-3 my-2 block">
        <Search
          size={12}
          className="pointer-events-none absolute start-2 top-1/2 -translate-y-1/2"
          style={{ color: 'var(--flow-text-dim)' }}
        />
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder={t('flow.calendar.unscheduled.search')}
          aria-label={t('flow.calendar.unscheduled.search')}
          className="w-full rounded-md py-1.5 pe-2 ps-7 text-flow-meta"
          style={{
            background: 'var(--flow-input-bg)',
            border: '1px solid var(--flow-border-control)',
            color: 'var(--flow-text)',
          }}
        />
      </label>
      <ul className="flex-1 space-y-1 overflow-auto px-3 pb-3">
        {shown.map((task) => {
          const id = String(task.id);
          const editable = draggable && canEdit(task);
          const color = tone(task);
          return (
            <li key={id}>
              <button
                type="button"
                title={task.title}
                draggable={editable}
                data-task={id}
                onDragStart={
                  editable
                    ? (event) => {
                        event.dataTransfer?.setData('text/plain', id);
                        if (event.dataTransfer) event.dataTransfer.effectAllowed = 'move';
                        scheduling.begin({ taskId: id, kind: 'schedule' });
                      }
                    : undefined
                }
                onDragEnd={editable ? scheduling.end : undefined}
                onClick={() => onOpen(task)}
                className={[
                  'flex w-full items-center gap-1.5 rounded-sm px-2 py-1.5 text-start text-flow-meta hover:brightness-95',
                  editable ? 'cursor-grab active:cursor-grabbing' : 'cursor-pointer',
                ].join(' ')}
                style={{
                  ...(color ? { ['--tone' as string]: color } : {}),
                  background: color ? withAlpha('var(--tone)', 13) : 'var(--flow-card)',
                  color: color ? onTint('var(--tone)') : 'var(--flow-text)',
                }}
              >
                <span dir="auto" className="truncate">
                  {task.title}
                </span>
              </button>
            </li>
          );
        })}
        {shown.length === 0 && (
          <li className="py-4 text-center text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {unscheduled.length === 0
              ? t('flow.calendar.unscheduled.empty')
              : t('flow.calendar.unscheduled.noMatches')}
          </li>
        )}
      </ul>
    </aside>
  );
}
