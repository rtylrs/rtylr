import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useMemo, useRef, useState } from 'react';
import { ClipboardList, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { getRouteApi } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { DatePicker } from '@voxire/ui';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import {
  addDays,
  dayKeyForDate,
  formatWeekDayShort,
  startOfWeek,
} from '../../shared/lib/dateUtils';
import { formatWorkEntryTarget } from '../../shared/lib/timeEntryLabel';
import { useTimeSettingsStore } from '../../shared/stores/timeSettingsStore';
import { useProjects } from '../projects/useProjects';
import { useTasks } from '../tasks/useTasks';
import AddHours from './AddHours';
import { todayLocalDate, useDeleteWorkDayLine, useWorkDay, useWorkDaysRange } from './useWorkDay';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

const routeApi = getRouteApi('/business/flow/time');

/**
 * My day: the screen people open every day to record what they worked on.
 */
const MyDay = (): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const { date: datePrefill } = routeApi.useSearch();
  const [date, setDate] = useState(datePrefill ?? todayLocalDate());

  const workDayQuery = useWorkDay(date);
  const { data, isLoading } = workDayQuery;
  const loadFailed = hasQueryFailed(workDayQuery);
  const deleteLine = useDeleteWorkDayLine();
  const addHoursRef = useRef<HTMLDivElement | null>(null);

  const weekStart = useTimeSettingsStore((s) => s.weekStart);
  const dailyTargetMinutes = useTimeSettingsStore((s) => s.dailyTargetMinutes);
  const weekDays = useMemo(() => {
    const start = startOfWeek(date, weekStart);
    return Array.from({ length: 7 }, (_, i) => addDays(start, i));
  }, [date, weekStart]);
  const { data: weekData } = useWorkDaysRange(weekDays[0], weekDays[6]);
  // The minutes, not a boolean. This was a Set of dates used only to decide
  // whether to draw a dot, so the strip carried the date and nothing else,
  // and the dot was aria-hidden with a title, which reads as nothing at all.
  const loggedByDate = useMemo(() => {
    const map = new Map<string, number>();
    for (const d of weekData?.workDays ?? []) {
      if (d.date) map.set(d.date, d.loggedMinutes ?? 0);
    }
    return map;
  }, [weekData]);

  const dayTargetMinutes = dailyTargetMinutes[dayKeyForDate(date)] ?? 0;
  const dayLoggedMinutes = data?.workDay?.loggedMinutes ?? 0;

  const lines = useMemo(() => data?.lines ?? [], [data]);

  // For labelling each line with what it was actually for - the same
  // project/task data Add Hours already fetches for its own pickers.
  const { data: projects } = useProjects();
  const { data: tasks } = useTasks({});
  const projectsById = useMemo(
    () => new Map((projects ?? []).map((project) => [project.id, project])),
    [projects]
  );
  const tasksById = useMemo(() => new Map((tasks ?? []).map((task) => [task.id, task])), [tasks]);

  return (
    <div className="h-full overflow-y-auto">
      <HelmetContainer title={`${t('flow.time.myDay')} · Flow`} />
      <div className="mx-auto max-w-[1180px] px-4 py-6 sm:px-6">
        <header className="mb-6">
          <h1 className="text-flow-screen font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.myDay')}
          </h1>
          {/* S-19: the day total belongs where the eye lands, at 20px. It was
              12px inside a pill at the top right of the Add-hours card, which
              is the last place you look on this screen. */}
          <p className="mt-1 flex items-baseline gap-1.5">
            <span
              className="text-flow-screen font-semibold tabular-nums"
              style={{ color: 'var(--flow-text)' }}
            >
              {formatDurationClock(dayLoggedMinutes)}
            </span>
            {dayTargetMinutes > 0 && (
              <span
                className="text-flow-body tabular-nums"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.of')} {formatDurationClock(dayTargetMinutes)}
              </span>
            )}
          </p>
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
              previousMonth: t('flowCommon.datePicker.previousMonth'),
              nextMonth: t('flowCommon.datePicker.nextMonth'),
            }}
          />
        </header>

        <div role="group" aria-label={t('flow.time.weekStrip')} className="mb-6 flex gap-1.5">
          {weekDays.map((d) => {
            const active = d === date;
            const dayMinutes = loggedByDate.get(d) ?? 0;
            return (
              <button
                key={d}
                type="button"
                onClick={() => setDate(d)}
                aria-current={active ? 'date' : undefined}
                className="flow-focus relative flex min-h-[44px] flex-1 flex-col items-center justify-center rounded-md py-1.5 text-flow-meta font-medium md:min-h-0"
                style={{
                  background: active ? 'var(--flow-active)' : 'var(--flow-hover)',
                  color: active ? 'var(--flow-on-accent)' : 'var(--flow-text)',
                }}
              >
                {formatWeekDayShort(d, i18n.language)}
                {/* The logged total under the date, in tabular digits. A dash
                    for a day with nothing on it, so the column keeps its
                    width and the row stays scannable. */}
                <span
                  className="mt-0.5 text-flow-meta tabular-nums"
                  style={{
                    color: active ? 'var(--flow-on-accent)' : 'var(--flow-text-muted)',
                    opacity: dayMinutes > 0 ? 1 : 0.55,
                  }}
                >
                  {dayMinutes > 0 ? formatDurationClock(dayMinutes) : '–'}
                </span>
              </button>
            );
          })}
        </div>

        {isLoading ? (
          <p className="text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.loadingDay')}
          </p>
        ) : loadFailed ? (
          // Without this the screen sat on "Loading the day..." forever,
          // because a failed query is not a loading one and nothing rendered
          // the difference.
          <ListLoadFailed onRetry={() => void workDayQuery.refetch()} />
        ) : (
          <>
            <section>
              <h2
                className="mb-3 text-flow-meta font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.whatYouWorkedOn')}
              </h2>

              {lines.length === 0 && (
                <div className="mb-3">
                  <EmptyState
                    icon={ClipboardList}
                    title={t('flow.time.noLinesYet')}
                    action={{
                      label: t('flow.time.noLinesYetCta'),
                      onClick: () =>
                        addHoursRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }),
                    }}
                    size="compact"
                  />
                </div>
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
                        className="text-flow-body font-semibold tabular-nums"
                        style={{ color: 'var(--flow-text)' }}
                      >
                        {formatDurationClock(line.durationMinutes ?? 0)}
                      </span>
                      <span
                        className="min-w-0 flex-1 text-flow-body"
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
                        className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm md:h-8 md:w-8"
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
              <div ref={addHoursRef}>
                <AddHours date={date} onDateChange={setDate} />
              </div>
            </section>
          </>
        )}
      </div>
    </div>
  );
};

export default MyDay;
