import React, { useMemo, useState } from 'react';
import { Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { getRouteApi } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { DatePicker } from '@voxire/ui';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import { addDays, formatWeekDayShort, startOfWeek } from '../../shared/lib/dateUtils';
import { formatWorkEntryTarget } from '../../shared/lib/timeEntryLabel';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTimeSettingsStore } from '../../shared/stores/timeSettingsStore';
import { useProjects } from '../projects/useProjects';
import { useTasks } from '../tasks/useTasks';
import AddHours from './AddHours';
import { todayLocalDate, useDeleteWorkDayLine, useWorkDay, useWorkDaysRange } from './useWorkDay';

const routeApi = getRouteApi('/time');

/**
 * My day: the screen people open every day to record what they worked on.
 */
const MyDay = (): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const { date: datePrefill } = routeApi.useSearch();
  const [date, setDate] = useState(datePrefill ?? todayLocalDate());

  const { data, isLoading } = useWorkDay(date);
  const deleteLine = useDeleteWorkDayLine();

  const weekStart = useTimeSettingsStore((s) => s.weekStart);
  const weekDays = useMemo(() => {
    const start = startOfWeek(date, weekStart);
    return Array.from({ length: 7 }, (_, i) => addDays(start, i));
  }, [date, weekStart]);
  const { data: weekData } = useWorkDaysRange(weekDays[0], weekDays[6]);
  const loggedDates = useMemo(() => {
    const set = new Set<string>();
    for (const d of weekData?.workDays ?? []) {
      if (d.date && (d.loggedMinutes ?? 0) > 0) set.add(d.date);
    }
    return set;
  }, [weekData]);

  const day = data?.workDay;
  const lines = useMemo(() => data?.lines ?? [], [data]);

  // For labelling each line with what it was actually for - the same
  // project/task data Add Hours already fetches for its own pickers.
  const activeWorkspaceId = useFlowStore((s) => s.activeWorkspaceId);
  const { data: projects } = useProjects(activeWorkspaceId);
  const { data: tasks } = useTasks({});
  const projectsById = useMemo(
    () => new Map((projects ?? []).map((project) => [project.id, project])),
    [projects]
  );
  const tasksById = useMemo(() => new Map((tasks ?? []).map((task) => [task.id, task])), [tasks]);

  const logged = day?.loggedMinutes ?? 0;

  return (
    <div className="h-full overflow-y-auto">
      <HelmetContainer title={`${t('flow.time.myDay')} · Flow`} />
      <div className="mx-auto max-w-[1180px] px-4 py-6 sm:px-6">
        <header className="mb-6">
          <h1 className="text-[20px] font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.myDay')}
          </h1>
          <label htmlFor="work-day-date" className="sr-only">
            {t('flow.time.date')}
          </label>
          <DatePicker
            id="work-day-date"
            value={date}
            onChange={setDate}
            locale={i18n.language}
            wrapperClassName="mt-2"
            className="w-auto"
            labels={{
              previousMonth: t('common.datePicker.previousMonth'),
              nextMonth: t('common.datePicker.nextMonth'),
            }}
          />
        </header>

        <div role="group" aria-label={t('flow.time.weekStrip')} className="mb-6 flex gap-1.5">
          {weekDays.map((d) => {
            const active = d === date;
            const hasEntries = loggedDates.has(d);
            return (
              <button
                key={d}
                type="button"
                onClick={() => setDate(d)}
                aria-current={active ? 'date' : undefined}
                className="flow-focus relative flex min-h-[44px] flex-1 flex-col items-center justify-center rounded-md py-1.5 text-[11px] font-medium md:min-h-0"
                style={{
                  background: active ? 'var(--accent)' : 'var(--flow-hover)',
                  color: active ? 'var(--flow-on-accent)' : 'var(--flow-text)',
                }}
              >
                {formatWeekDayShort(d, i18n.language)}
                {hasEntries && (
                  <span
                    title={t('flow.time.weekDayLogged')}
                    aria-hidden="true"
                    className="absolute bottom-0.5 h-1 w-1 rounded-full"
                    style={{ background: active ? 'var(--flow-on-accent)' : 'var(--accent)' }}
                  />
                )}
              </button>
            );
          })}
        </div>

        {isLoading ? (
          <p className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.loadingDay')}
          </p>
        ) : (
          <>
            <section
              className="mb-6 flex flex-wrap items-baseline gap-x-6 gap-y-2 rounded-md p-3"
              style={{ background: 'var(--flow-hover)' }}
              aria-live="polite"
            >
              <Figure label={t('flow.time.logged')} value={formatDurationClock(logged)} />
            </section>

            <section>
              <h2
                className="mb-3 text-[11px] font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.whatYouWorkedOn')}
              </h2>

              {lines.length === 0 && (
                <p className="mb-3 text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.time.noLinesYet')}
                </p>
              )}

              <ul className="mb-4 space-y-1.5">
                {lines.map((line) => {
                  const project = line.projectId ? projectsById.get(line.projectId) : undefined;
                  const task = line.taskId ? tasksById.get(line.taskId) : undefined;
                  const target = formatWorkEntryTarget(task, project);
                  return (
                    <li
                      key={line.id}
                      className="flex items-start gap-3 rounded-md p-2"
                      style={{ background: 'var(--flow-hover)' }}
                    >
                      <span
                        className="text-[13px] font-semibold tabular-nums"
                        style={{ color: 'var(--accent-fg)' }}
                      >
                        {formatDurationClock(line.durationMinutes ?? 0)}
                      </span>
                      <span
                        className="min-w-0 flex-1 text-[13px]"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {target && <span className="block font-semibold">{target}</span>}
                        {/* Add Hours submits "-" for an empty note (the wire
                         * requires a non-empty description) - showing that
                         * placeholder here would read as a stray artifact
                         * under the task name, not as "no note was left". */}
                        {line.description && line.description !== '-' && (
                          <span style={{ color: target ? 'var(--flow-text-muted)' : undefined }}>
                            {line.description}
                          </span>
                        )}
                      </span>
                      <button
                        type="button"
                        onClick={() => line.id && deleteLine.mutate({ id: line.id, date })}
                        aria-label={t('flow.time.deleteLineLabel', {
                          duration: formatDurationClock(line.durationMinutes ?? 0),
                        })}
                        className="flow-focus flex h-11 w-11 shrink-0 items-center justify-center rounded md:h-8 md:w-8"
                        style={{ color: 'var(--flow-text-muted)' }}
                      >
                        <Trash2 size={14} aria-hidden="true" />
                      </button>
                    </li>
                  );
                })}
              </ul>

              {/* The old "+ Log time" button opened a modal (LogTimeDialog);
               * that dialog and its own day-must-exist gate are gone. Add
               * Hours renders inline instead, and creates the work day itself
               * on first submit if one doesn't exist yet for this date. */}
              <AddHours date={date} onDateChange={setDate} />
            </section>
          </>
        )}
      </div>
    </div>
  );
};

const Figure = ({
  label,
  value,
  tone,
}: {
  label: string;
  value: string;
  tone?: string;
}): React.JSX.Element => (
  <div>
    <span
      className="block text-[10px] font-semibold uppercase tracking-wider"
      style={{ color: 'var(--flow-text-muted)' }}
    >
      {label}
    </span>
    <span
      className="text-[18px] font-semibold tabular-nums"
      style={{ color: tone ?? 'var(--flow-text)' }}
    >
      {value}
    </span>
  </div>
);

export default MyDay;
