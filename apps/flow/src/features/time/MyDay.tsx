import React, { useMemo, useState } from 'react';
import { Plus, Trash2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { formatDurationClock, parseDurationInput } from '../../shared/lib/durationUtils';
import { addDays, formatWeekDayShort, startOfWeek } from '../../shared/lib/dateUtils';
import DurationField from '../tasks/components/DurationField';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useProjects } from '../projects/useProjects';
import { useTimeSettingsStore } from '../../shared/stores/timeSettingsStore';
import {
  newClientKey,
  todayLocalDate,
  useCreateWorkDayLine,
  useDeleteWorkDayLine,
  useUpsertWorkDay,
  useWorkDay,
  useWorkDaysRange,
  useWorkTypes,
} from './useWorkDay';

const WORK_MODES = ['onsite', 'remote', 'hybrid'] as const;

const labelClass = 'mb-1 block text-[10px] font-semibold uppercase tracking-wider';
const inputStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border)',
  color: 'var(--flow-text)',
};

/**
 * My day: the screen people open every day to record what they worked on.
 *
 * Two levels, because a working day contains several pieces of work: the day
 * carries the mode and the clock, and its lines carry the work. The gap between
 * the two is shown and never blocks, because a blocked submission produces
 * invented lines rather than honest ones.
 */
const MyDay = (): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const [date, setDate] = useState(todayLocalDate());

  const { data, isLoading } = useWorkDay(date);
  const { data: workTypesData } = useWorkTypes();
  const activeWorkspaceId = useFlowStore((s) => s.activeWorkspaceId);
  const { data: projects } = useProjects(activeWorkspaceId);
  const upsertDay = useUpsertWorkDay();
  const createLine = useCreateWorkDayLine();
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
  const workTypes = workTypesData?.workTypes ?? [];

  const [duration, setDuration] = useState('');
  const [description, setDescription] = useState('');
  const [projectId, setProjectId] = useState('');
  const [workTypeId, setWorkTypeId] = useState('');

  const clocked = day?.clockedMinutes ?? 0;
  const logged = day?.loggedMinutes ?? 0;
  const gap = day?.gapMinutes ?? 0;

  const saveDay = (
    patch: Partial<{ workMode: string; startTime: string; endTime: string; breakMinutes: number }>
  ) => {
    upsertDay.mutate({
      date,
      workMode: patch.workMode ?? day?.workMode ?? 'onsite',
      startTime: patch.startTime ?? day?.startTime ?? undefined,
      endTime: patch.endTime ?? day?.endTime ?? undefined,
      breakMinutes: patch.breakMinutes ?? day?.breakMinutes ?? 0,
    });
  };

  const canAddLine = parseDurationInput(duration) !== null && projectId !== '';

  const addLine = () => {
    const minutes = parseDurationInput(duration);
    if (minutes === null || !day?.id || !projectId) return;
    createLine.mutate(
      {
        date,
        workDayId: day.id,
        // Generated here, before the request leaves, so a retry after a dropped
        // connection resolves to the same entry rather than a second one.
        clientKey: newClientKey(),
        durationMinutes: minutes,
        description: description.trim(),
        projectId,
        workTypeId: workTypeId || undefined,
        isBillable: true,
      },
      {
        onSuccess: () => {
          setDuration('');
          setDescription('');
          setProjectId('');
        },
      }
    );
  };

  return (
    <div className="h-full overflow-y-auto">
      <HelmetContainer title="My day · rtylr" />
      <div className="mx-auto max-w-3xl px-4 py-6 sm:px-6">
        <header className="mb-6">
          <h1 className="text-[20px] font-bold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.myDay')}
          </h1>
          <label htmlFor="work-day-date" className="sr-only">
            {t('flow.time.date')}
          </label>
          <input
            id="work-day-date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="flow-focus mt-2 min-h-[44px] rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
            style={inputStyle}
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
            <section className="mb-6">
              <h2
                className="mb-3 text-[11px] font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.theDay')}
              </h2>

              <div className="mb-4">
                <span className={labelClass} style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.time.workMode')}
                </span>
                <div
                  role="radiogroup"
                  aria-label={t('flow.time.workMode')}
                  className="flex flex-wrap gap-2"
                >
                  {WORK_MODES.map((mode) => {
                    const active = (day?.workMode ?? 'onsite') === mode;
                    return (
                      <button
                        key={mode}
                        type="button"
                        role="radio"
                        aria-checked={active}
                        onClick={() => saveDay({ workMode: mode })}
                        className="flow-focus min-h-[44px] rounded-md px-3 text-[13px] font-medium md:min-h-0 md:py-2"
                        style={{
                          background: active ? 'var(--accent)' : 'var(--flow-hover)',
                          color: active ? 'var(--flow-on-accent)' : 'var(--flow-text-muted)',
                        }}
                      >
                        {t(`flow.time.mode.${mode}`)}
                      </button>
                    );
                  })}
                </div>
              </div>

              <div className="grid gap-3 sm:grid-cols-3">
                <div>
                  <label
                    htmlFor="day-start"
                    className={labelClass}
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {t('flow.time.startTime')}
                  </label>
                  <input
                    id="day-start"
                    type="time"
                    defaultValue={day?.startTime ?? ''}
                    onBlur={(e) => saveDay({ startTime: e.target.value })}
                    className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                    style={inputStyle}
                  />
                </div>
                <div>
                  <label
                    htmlFor="day-end"
                    className={labelClass}
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {t('flow.time.endTime')}
                  </label>
                  <input
                    id="day-end"
                    type="time"
                    defaultValue={day?.endTime ?? ''}
                    onBlur={(e) => saveDay({ endTime: e.target.value })}
                    className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                    style={inputStyle}
                  />
                </div>
                <div>
                  <label
                    htmlFor="day-break"
                    className={labelClass}
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {t('flow.time.breakMinutes')}
                  </label>
                  <input
                    id="day-break"
                    type="number"
                    min={0}
                    defaultValue={day?.breakMinutes ?? 0}
                    onBlur={(e) => saveDay({ breakMinutes: Number(e.target.value) || 0 })}
                    className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                    style={inputStyle}
                  />
                </div>
              </div>
            </section>

            {/* Clocked against logged. The gap is the most useful number here, so
             * it is shown plainly and never blocks: enforcing it produces
             * invented lines rather than an honest total. */}
            <section
              className="mb-6 flex flex-wrap items-baseline gap-x-6 gap-y-2 rounded-md p-3"
              style={{ background: 'var(--flow-hover)' }}
              aria-live="polite"
            >
              <Figure label={t('flow.time.clocked')} value={formatDurationClock(clocked)} />
              <Figure label={t('flow.time.logged')} value={formatDurationClock(logged)} />
              <Figure
                label={t('flow.time.unaccounted')}
                value={formatDurationClock(Math.abs(gap))}
                tone={gap === 0 ? undefined : 'var(--flow-status-pending)'}
              />
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
                {lines.map((line) => (
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
                      {line.description}
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
                ))}
              </ul>

              {/* The day has to exist before work can be filed under it, which is
               * what the upsert above creates. */}
              {!day?.id ? (
                <p className="text-[13px]" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.time.setTheDayFirst')}
                </p>
              ) : (
                <div className="grid gap-3 sm:grid-cols-[10rem_1fr_auto] sm:items-end">
                  <DurationField
                    id="line-duration"
                    label={t('flow.time.duration')}
                    value={duration}
                    onChange={setDuration}
                    onSubmit={addLine}
                    placeholder={t('flow.time.durationPlaceholder')}
                    invalidMessage={t('flow.time.durationInvalid')}
                  />
                  <div>
                    <label
                      htmlFor="line-description"
                      className={labelClass}
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {t('flow.time.description')}
                    </label>
                    <input
                      id="line-description"
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && addLine()}
                      className="flow-focus min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                      style={inputStyle}
                    />
                    <select
                      aria-label={t('flow.time.project')}
                      value={projectId}
                      onChange={(e) => setProjectId(e.target.value)}
                      required
                      className="flow-focus mt-2 min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                      style={inputStyle}
                    >
                      <option value="">{t('flow.time.selectProject')}</option>
                      {(projects ?? []).map((project) => (
                        <option key={project.id} value={project.id}>
                          {project.name}
                        </option>
                      ))}
                    </select>
                    {workTypes.length > 0 && (
                      <select
                        aria-label={t('flow.time.workType')}
                        value={workTypeId}
                        onChange={(e) => setWorkTypeId(e.target.value)}
                        className="flow-focus mt-2 min-h-[44px] w-full rounded px-2 text-[16px] md:min-h-0 md:py-1.5 md:text-[13px]"
                        style={inputStyle}
                      >
                        <option value="">{t('flow.time.noWorkType')}</option>
                        {workTypes.map((type) => (
                          <option key={type.id} value={type.id}>
                            {type.name}
                          </option>
                        ))}
                      </select>
                    )}
                  </div>
                  <button
                    type="button"
                    onClick={addLine}
                    disabled={!canAddLine || createLine.isPending}
                    className="flow-focus inline-flex min-h-[44px] items-center justify-center gap-1.5 rounded px-4 text-[13px] font-semibold disabled:opacity-40 md:min-h-0 md:py-2.5"
                    style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
                  >
                    <Plus size={14} aria-hidden="true" />
                    {t('flow.time.addLine')}
                  </button>
                </div>
              )}
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
