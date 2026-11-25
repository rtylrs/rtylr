import React, { useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import FlowSelect from '../../shared/components/FlowSelect';
import GateNotice from '../../shared/components/GateNotice';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { addDays, startOfWeek, todayIso } from '../../shared/lib/dateUtils';
import { formatShortDate } from '../../shared/lib/utils';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { withAlpha } from '../../shared/lib/colorUtils';
import { usePlannedWorkload } from './useWorkDay';
import {
  WORK_WEEKS,
  buildPlannedRows,
  formatHours,
  type CellTone,
  type WorkWeek,
} from './plannedModel';

const DAYS_SHOWN = 14;
const PREFS_KEY = 'flow.teamHours.planned';

const TONE: Record<CellTone, { background: string; color: string }> = {
  off: { background: 'transparent', color: 'var(--flow-text-dim)' },
  under: { background: withAlpha('var(--flow-warn-fg)', 24), color: 'var(--flow-text)' },
  ok: { background: withAlpha('var(--color-success)', 22), color: 'var(--flow-text)' },
  over: { background: withAlpha('var(--flow-danger-fg)', 30), color: 'var(--flow-text)' },
};

interface Prefs {
  workWeek: WorkWeek;
  targetHours: number;
}

// Remembered in this browser only: they shape the view, not the data.
function loadPrefs(): Prefs {
  try {
    const raw = JSON.parse(localStorage.getItem(PREFS_KEY) ?? '{}') as Partial<Prefs>;
    return {
      workWeek: raw.workWeek === 'sun_thu' ? 'sun_thu' : 'mon_fri',
      targetHours:
        typeof raw.targetHours === 'number' && raw.targetHours > 0 && raw.targetHours <= 24
          ? raw.targetHours
          : 5,
    };
  } catch {
    return { workWeek: 'mon_fri', targetHours: 5 };
  }
}

function savePrefs(prefs: Prefs): void {
  try {
    localStorage.setItem(PREFS_KEY, JSON.stringify(prefs));
  } catch {
    // Private mode or blocked storage: the view still works for this visit.
  }
}

/**
 * Team hours > Planned (T-341): planned hours per person per day from task
 * estimates, spread over each task's start-to-due working days, against a
 * daily target. A cell opens the tasks behind it.
 */
const PlannedWorkload: React.FC<{ team: ReadonlyArray<{ id: string; name: string }> }> = ({
  team,
}) => {
  const { t, i18n } = useTranslation();
  const openTask = useOpenTask();
  const [prefs, setPrefsState] = useState<Prefs>(loadPrefs);
  const setPrefs = (next: Prefs) => {
    setPrefsState(next);
    savePrefs(next);
  };
  const [start, setStart] = useState(() =>
    startOfWeek(todayIso(), prefs.workWeek === 'sun_thu' ? 'sunday' : 'monday')
  );
  const [selected, setSelected] = useState<{ userId: string; date: string } | null>(null);
  const dates = useMemo(
    () => Array.from({ length: DAYS_SHOWN }, (_, i) => addDays(start, i)),
    [start]
  );
  const query = usePlannedWorkload(
    dates[0],
    dates[dates.length - 1],
    WORK_WEEKS[prefs.workWeek].iso
  );
  const targetMinutes = Math.round(prefs.targetHours * 60);
  const rows = useMemo(
    () => buildPlannedRows(query.data, team, dates, prefs.workWeek, targetMinutes),
    [query.data, team, dates, prefs.workWeek, targetMinutes]
  );
  const tasksById = useMemo(
    () => new Map((query.data?.tasks ?? []).map((task) => [String(task.id), task])),
    [query.data]
  );
  const selectedCell = selected
    ? rows.find((r) => r.userId === selected.userId)?.cells.find((c) => c.date === selected.date)
    : undefined;
  const selectedName = rows.find((r) => r.userId === selected?.userId)?.name ?? '';

  const failed = hasQueryFailed(query);
  const loading = query.isPending && !failed;

  return (
    <section aria-labelledby="planned-heading" className="space-y-4">
      <div className="flex flex-wrap items-end gap-3">
        <h2 id="planned-heading" className="sr-only">
          {t('flow.time.planned.title')}
        </h2>
        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => setStart(addDays(start, -7))}
            aria-label={t('flow.time.planned.previousWeek')}
            className="flow-focus h-8 w-8 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <ChevronLeft size={14} className="rtl:rotate-180" aria-hidden="true" />
          </button>
          <span className="text-flow-body font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.planned.range', {
              start: formatShortDate(dates[0], i18n.language),
              end: formatShortDate(dates[dates.length - 1], i18n.language),
            })}
          </span>
          <button
            type="button"
            onClick={() => setStart(addDays(start, 7))}
            aria-label={t('flow.time.planned.nextWeek')}
            className="flow-focus h-8 w-8 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <ChevronRight size={14} className="rtl:rotate-180" aria-hidden="true" />
          </button>
        </div>
        <div className="w-[170px]">
          <FlowSelect
            label={t('flow.time.planned.workWeek')}
            value={prefs.workWeek}
            options={(['mon_fri', 'sun_thu'] as const).map((value) => ({
              value,
              label: t(`flow.time.planned.workWeeks.${value}`),
            }))}
            onChange={(value) => setPrefs({ ...prefs, workWeek: value as WorkWeek })}
          />
        </div>
        <label
          className="flex items-center gap-2 text-flow-meta"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.time.planned.target')}
          <input
            type="number"
            min={0.5}
            max={24}
            step={0.5}
            value={prefs.targetHours}
            onChange={(e) => {
              const value = Number(e.target.value);
              if (value > 0 && value <= 24) setPrefs({ ...prefs, targetHours: value });
            }}
            className="flow-focus w-16 rounded-md px-2 py-1 text-flow-meta tabular-nums"
            style={{
              background: 'var(--flow-input-bg)',
              border: '1px solid var(--flow-border-control)',
              color: 'var(--flow-text)',
            }}
          />
        </label>
      </div>
      <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
        {t('flow.time.planned.explain', { target: prefs.targetHours })}
      </p>

      {failed ? (
        <ListLoadFailed onRetry={() => void query.refetch()} />
      ) : loading ? (
        <p className="text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.time.loadingDay')}
        </p>
      ) : rows.length === 0 ? (
        <GateNotice reason={t('flow.time.planned.empty')} className="text-flow-body" />
      ) : (
        <div
          className="overflow-x-auto rounded-md"
          style={{ border: '1px solid var(--flow-border)' }}
        >
          <table className="w-full border-collapse text-flow-meta">
            <thead>
              <tr style={{ color: 'var(--flow-text-dim)' }}>
                <th
                  scope="col"
                  className="sticky start-0 z-[1] min-w-[150px] px-3 py-2 text-start font-semibold"
                  style={{ background: 'var(--flow-surface)' }}
                >
                  {t('flow.time.planned.person')}
                </th>
                {dates.map((date) => {
                  const d = new Date(`${date}T00:00:00`);
                  return (
                    <th
                      key={date}
                      scope="col"
                      className="min-w-[44px] px-1 py-2 text-center font-semibold"
                      style={{ color: date === todayIso() ? 'var(--accent)' : undefined }}
                    >
                      <span className="block">
                        {d.toLocaleDateString(i18n.language, { weekday: 'short' })}
                      </span>
                      <span className="block">{d.getDate()}</span>
                    </th>
                  );
                })}
                <th scope="col" className="px-3 py-2 text-end font-semibold">
                  {t('flow.time.planned.total')}
                </th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr key={row.userId} style={{ borderTop: '1px solid var(--flow-border)' }}>
                  <th
                    scope="row"
                    className="sticky start-0 z-[1] px-3 py-1.5 text-start font-medium truncate"
                    style={{ background: 'var(--flow-surface)', color: 'var(--flow-text)' }}
                  >
                    {row.name}
                  </th>
                  {row.cells.map((cell) => {
                    const isSelected =
                      selected?.userId === row.userId && selected.date === cell.date;
                    const tone = TONE[cell.tone];
                    return (
                      <td key={cell.date} className="p-0.5 text-center">
                        <button
                          type="button"
                          disabled={cell.minutes === 0}
                          onClick={() =>
                            setSelected(isSelected ? null : { userId: row.userId, date: cell.date })
                          }
                          aria-pressed={isSelected}
                          aria-label={t('flow.time.planned.cellLabel', {
                            name: row.name,
                            date: formatShortDate(cell.date, i18n.language),
                            hours: formatHours(cell.minutes),
                            status:
                              cell.tone === 'off'
                                ? ''
                                : t(`flow.time.planned.legend.${cell.tone}`, {
                                    target: prefs.targetHours,
                                  }),
                          })}
                          title={
                            cell.tone === 'off'
                              ? undefined
                              : t(`flow.time.planned.legend.${cell.tone}`, {
                                  target: prefs.targetHours,
                                })
                          }
                          className="flow-focus w-full rounded-sm py-1.5 tabular-nums disabled:cursor-default"
                          style={{
                            background: tone.background,
                            color: tone.color,
                            outline: isSelected ? '2px solid var(--accent)' : undefined,
                          }}
                        >
                          {cell.minutes > 0 ? formatHours(cell.minutes) : cell.working ? '0' : ''}
                        </button>
                      </td>
                    );
                  })}
                  <td
                    className="px-3 py-1.5 text-end font-semibold tabular-nums"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {formatHours(row.totalMinutes)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div
        className="flex flex-wrap gap-3 text-flow-meta"
        style={{ color: 'var(--flow-text-muted)' }}
      >
        {(['under', 'ok', 'over'] as const).map((tone) => (
          <span key={tone} className="flex items-center gap-1.5">
            <span className="h-3 w-3 rounded-sm" style={{ background: TONE[tone].background }} />
            {t(`flow.time.planned.legend.${tone}`, { target: prefs.targetHours })}
          </span>
        ))}
      </div>

      {selectedCell && selectedCell.taskIds.length > 0 && (
        <div
          className="rounded-md p-3 space-y-1.5"
          style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
        >
          <h3 className="text-flow-body font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.planned.cellTasks', {
              name: selectedName,
              date: formatShortDate(selectedCell.date, i18n.language),
              hours: formatHours(selectedCell.minutes),
            })}
          </h3>
          <ul className="space-y-1">
            {selectedCell.taskIds.map((id) => {
              const task = tasksById.get(id);
              return (
                <li key={id}>
                  <button
                    type="button"
                    onClick={() => openTask(id)}
                    className="flow-focus flex w-full items-baseline gap-2 text-start text-flow-meta hover:underline"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    <span style={{ color: 'var(--flow-text-dim)' }}>{task?.identifier ?? id}</span>
                    <span className="truncate" dir="auto">
                      {task?.title ?? ''}
                    </span>
                    {task && (
                      <span className="ms-auto shrink-0" style={{ color: 'var(--flow-text-dim)' }}>
                        {t('flow.time.planned.taskSpan', {
                          hours: formatHours(task.estimateMinutes ?? 0),
                          start: task.startDate
                            ? formatShortDate(task.startDate, i18n.language)
                            : formatShortDate(task.dueDate ?? '', i18n.language),
                          end: formatShortDate(task.dueDate || task.startDate || '', i18n.language),
                        })}
                      </span>
                    )}
                  </button>
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </section>
  );
};

export default PlannedWorkload;
