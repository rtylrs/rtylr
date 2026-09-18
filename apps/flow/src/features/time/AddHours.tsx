import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, getRouteApi } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { CircleAlert } from 'lucide-react';
import { DatePicker } from '@voxire/ui';
import { formatDurationClock } from '../../shared/lib/durationUtils';
import {
  dayKeyForDate,
  formatFullDayLabel,
  formatEntryTimestamp,
} from '../../shared/lib/dateUtils';
import { formatWorkEntryTarget } from '../../shared/lib/timeEntryLabel';
import { useTimeSettingsStore } from '../../shared/stores/timeSettingsStore';
import { useProjects } from '../projects/useProjects';
import { useTasks } from '../tasks/useTasks';
import {
  newClientKey,
  useCreateWorkDayLine,
  useDeleteWorkDayLine,
  useUpsertWorkDay,
  useWorkDay,
  type WorkDay,
} from './useWorkDay';
import { WORK_MODES as LOCATIONS, type WorkMode as Location } from './workModes';

const routeApi = getRouteApi('/time');

const labelClass = 'mb-[7px] block text-meta font-bold uppercase tracking-[0.07em]' as const;
const fieldStyle: React.CSSProperties = {
  background: 'var(--flow-input-bg)',
  border: '1px solid var(--flow-border-control)',
  color: 'var(--flow-text)',
};
const fieldSizing =
  'min-h-[44px] w-full rounded-lg px-3 text-title md:h-[38px] md:min-h-0 md:text-meta';

type Phase = 'idle' | 'submitting' | 'error' | 'success';

interface SuccessInfo {
  durationMinutes: number;
  target: string;
  dayLabel: string;
  newTotalMinutes: number;
}

interface AddHoursProps {
  date: string;
  onDateChange: (date: string) => void;
}

function randomReference(): string {
  return `TL-${Math.random().toString(36).slice(2, 6).toUpperCase()}`;
}

const LAST_LOCATION_STORAGE_KEY = 'flow.addHours.lastLocation';

/** The location from the person's last submitted entry, in this browser - a
 * plain convenience default, not a setting, so it needs no seam disclosure:
 * nothing server-side is ever decided from it, only which button starts
 * selected on the next entry. */
function getLastUsedLocation(): Location | null {
  try {
    const stored = localStorage.getItem(LAST_LOCATION_STORAGE_KEY);
    return (LOCATIONS as readonly string[]).includes(stored ?? '') ? (stored as Location) : null;
  } catch {
    return null;
  }
}

function rememberLastUsedLocation(location: Location): void {
  try {
    localStorage.setItem(LAST_LOCATION_STORAGE_KEY, location);
  } catch {
    // Best-effort only - a full storage quota or disabled storage shouldn't
    // block logging hours.
  }
}

/**
 * The inline form that replaced the "+ Log time" dialog on My day. One entry,
 * one commit per submit (Log another is the loop for a second entry) - no
 * draft state, no approval step here: review and approval happen downstream
 * on the week. No timer either; the timer that starts from a task lives in
 * the shell bar, not this screen.
 *
 * Keyed by `date` from MyDay, so switching the day strip is a full remount:
 * every field, the banner and the phase reset together. "Log another" (the
 * success banner's own action) intentionally does not go through that reset -
 * it clears only duration and note, in place, because a second entry usually
 * repeats the same project, task and location.
 */
const AddHours = ({ date, onDateChange }: AddHoursProps): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const search = routeApi.useSearch();
  const dailyTargetMinutes = useTimeSettingsStore((s) => s.dailyTargetMinutes);

  const { data } = useWorkDay(date);
  const day: WorkDay | undefined = data?.workDay;

  // Each calendar day has its own configured target (Settings > Time), not a
  // weekly figure divided evenly - a non-working day is genuinely 0, not a
  // fifth of the week. See the settings-store seam pattern in
  // shared/stores/timeSettingsStore.ts.
  const dayTargetMinutes = dailyTargetMinutes[dayKeyForDate(date)] ?? 0;
  const dayLoggedMinutes = day?.loggedMinutes ?? 0;

  // The only real lock signal available is this day's own server-reported
  // status. The local time-settings store's approvalLocksEntry/autoLockDays
  // are per-browser shipped defaults with no endpoint yet (see that file's
  // own header comment) - deciding whether a day is locked from them would be
  // the UI asserting a server decision it cannot know. `status` is real,
  // server-populated data on this exact WorkDay, so it is what gates this.
  const isDayLocked = day?.status === 'approved';

  const [hours, setHours] = useState('0');
  const [minutesRaw, setMinutesRaw] = useState('00');
  const [minutesTouched, setMinutesTouched] = useState(false);
  // Defaults to today's own recorded location if this day already has one,
  // else to whatever location this person's last submitted entry used (any
  // day) - a hybrid day genuinely can have some entries logged differently,
  // but repeating the previous pick is right far more often than 'onsite' is.
  const [location, setLocation] = useState<Location>(() => {
    const fromDay = day?.workMode;
    if ((LOCATIONS as readonly string[]).includes(fromDay ?? '')) return fromDay as Location;
    return getLastUsedLocation() ?? 'onsite';
  });
  const [projectId, setProjectId] = useState(search.project ?? '');
  const [taskId, setTaskId] = useState(search.task ?? '');
  const [note, setNote] = useState('');
  const [phase, setPhase] = useState<Phase>('idle');
  const [errorRef, setErrorRef] = useState('');
  const [successInfo, setSuccessInfo] = useState<SuccessInfo | null>(null);
  const [lastLineId, setLastLineId] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const bannerRef = useRef<HTMLDivElement>(null);
  const copiedTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => () => clearTimeout(copiedTimeoutRef.current), []);

  useEffect(() => {
    if (phase === 'error' || phase === 'success') bannerRef.current?.focus();
  }, [phase]);

  const { data: projects } = useProjects();
  const { data: tasks } = useTasks({ projectId: projectId || undefined });

  const upsertDay = useUpsertWorkDay({ silent: true });
  const createLine = useCreateWorkDayLine({ silent: true });
  const deleteLine = useDeleteWorkDayLine();

  const selectedProject = useMemo(
    () => (projects ?? []).find((p) => p.id === projectId),
    [projects, projectId]
  );
  const selectedTask = useMemo(
    () => (tasks ?? []).find((task) => task.id === taskId),
    [tasks, taskId]
  );
  const taskDescription = selectedTask?.description?.trim() ?? '';

  const targetLabel = useMemo(
    () => formatWorkEntryTarget(selectedTask, selectedProject),
    [selectedTask, selectedProject]
  );

  const hoursInt = Number.parseInt(hours, 10) || 0;
  const minutesInt = Number.parseInt(minutesRaw, 10) || 0;
  const durationMinutes = hoursInt * 60 + minutesInt;

  const minutesError =
    minutesTouched && minutesInt > 59
      ? t('flow.time.add.errors.minutesRange', {
          hint: t('flow.time.add.errors.minutesHint', {
            count: Math.floor(minutesInt / 60),
            mins: minutesInt % 60,
          }),
        })
      : '';

  const dateError = isDayLocked
    ? day?.approvedAt
      ? t('flow.time.add.errors.dayLocked', {
          date: formatEntryTimestamp(day.approvedAt, i18n.language),
        })
      : t('flow.time.add.errors.dayLockedGeneric')
    : '';

  const projectedTotalMinutes = dayLoggedMinutes + durationMinutes;
  const showsOverCapacityWarning = durationMinutes > 0 && projectedTotalMinutes > dayTargetMinutes;

  const disabledReason = isDayLocked
    ? dateError
    : durationMinutes <= 0
      ? t('flow.time.add.reasonNoDuration')
      : !projectId
        ? t('flow.time.add.reasonNoProject')
        : !taskId
          ? t('flow.time.add.reasonNoTask')
          : minutesError
            ? minutesError
            : '';

  const canSubmit = !disabledReason;
  // Submitting makes every field read-only, never disabled: the person must
  // still be able to see (and, for the note, copy) what they typed if the
  // request fails. Native <select> and <button> have no readOnly attribute,
  // so this guards their handlers instead of graying the control out.
  const locked = phase === 'submitting';
  // Fields are left populated after a success (state 5's own "Log another"
  // keeps date/project/task/location on purpose - see resetForLogAnother).
  // Without this, the still-filled, still-enabled form invites clicking
  // "Log hours" a second time on data that already logged, silently filing
  // a duplicate entry. Success only clears via the banner's own actions.
  const submitBlocked = locked || phase === 'success';

  const resetForLogAnother = useCallback((): void => {
    setHours('0');
    setMinutesRaw('00');
    setMinutesTouched(false);
    setNote('');
    setPhase('idle');
    setSuccessInfo(null);
    setLastLineId(null);
  }, []);

  const handleSubmit = useCallback(async (): Promise<void> => {
    if (!canSubmit || submitBlocked) return;
    setPhase('submitting');
    try {
      let workDayId = day?.id;
      if (!workDayId) {
        const upserted = await upsertDay.mutateAsync({
          date,
          workMode: day?.workMode ?? location,
          breakMinutes: day?.breakMinutes ?? 0,
        });
        workDayId = upserted.workDay?.id;
      }
      if (!workDayId) throw new Error(t('errors.noWorkDay'));

      const result = await createLine.mutateAsync({
        date,
        workDayId,
        clientKey: newClientKey(),
        durationMinutes,
        // The wire requires a non-empty description even though this field
        // is optional in this form - same substitution already used for the
        // legacy TimeLog model's own required-but-optional description (see
        // docs/flow-time-tracking.md).
        description: note.trim() || '-',
        projectId,
        taskId,
        workMode: location,
        isBillable: true,
      });

      setLastLineId(result.line?.id ?? null);
      rememberLastUsedLocation(location);
      setSuccessInfo({
        durationMinutes,
        target: targetLabel,
        dayLabel: formatFullDayLabel(date, i18n.language),
        newTotalMinutes: projectedTotalMinutes,
      });
      setPhase('success');
    } catch {
      setErrorRef(randomReference());
      setPhase('error');
    }
  }, [
    canSubmit,
    submitBlocked,
    day,
    upsertDay,
    date,
    location,
    createLine,
    durationMinutes,
    note,
    projectId,
    taskId,
    targetLabel,
    i18n.language,
    projectedTotalMinutes,
    t,
  ]);

  const handleDiscard = useCallback((): void => {
    setHours('0');
    setMinutesRaw('00');
    setMinutesTouched(false);
    setProjectId('');
    setTaskId('');
    setNote('');
    setPhase('idle');
  }, []);

  const handleUndo = useCallback((): void => {
    if (!lastLineId) return;
    deleteLine.mutate(
      { id: lastLineId, date },
      {
        onSuccess: () => {
          setPhase('idle');
          setSuccessInfo(null);
          setLastLineId(null);
        },
      }
    );
  }, [lastLineId, date, deleteLine]);

  const handleCopyNote = useCallback((): void => {
    if (!navigator.clipboard) return;
    void navigator.clipboard.writeText(note).then(() => {
      setCopied(true);
      clearTimeout(copiedTimeoutRef.current);
      copiedTimeoutRef.current = setTimeout(() => setCopied(false), 2000);
    });
  }, [note]);

  const handleFormKeyDown = (e: React.KeyboardEvent<HTMLDivElement>): void => {
    const target = e.target as HTMLElement;
    if (e.key === 'Enter' && target.tagName !== 'TEXTAREA') {
      e.preventDefault();
      void handleSubmit();
    } else if (e.key === 'Escape') {
      handleDiscard();
    }
  };

  const dayLabel = formatFullDayLabel(date, i18n.language);
  const noteMax = 300;
  const noteDangerAt = 280;

  return (
    <div
      className="w-full max-w-[1180px] rounded-xl"
      style={{
        background: 'var(--flow-surface)',
        border: '1px solid var(--flow-border)',
        boxShadow: 'var(--flow-shadow)',
      }}
      onKeyDown={handleFormKeyDown}
    >
      {/* Header. My day's own totals strip only ever shows Logged now (no
       * target comparison lives there), so this pill is the one place that
       * shows today's progress against its configured target. */}
      <div
        className="flex flex-col gap-2 rounded-t-xl px-6 py-4 md:flex-row md:items-center md:justify-between"
        style={{ background: 'var(--flow-card)', borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h3 className="text-section font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.add.title')}
          </h3>
          <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
            {t('flow.time.add.subtitle', { day: dayLabel })}
          </p>
        </div>
        <div
          className="flex h-8 items-center gap-1 self-start rounded-full px-3 text-meta md:self-auto"
          style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
        >
          <span>{t('flow.time.add.today')}</span>
          <strong className="tabular-nums font-semibold" style={{ color: 'var(--flow-text)' }}>
            {formatDurationClock(dayLoggedMinutes)}
          </strong>
          <span>{t('flow.time.of')}</span>
          <strong className="tabular-nums font-semibold" style={{ color: 'var(--flow-text)' }}>
            {formatDurationClock(dayTargetMinutes)}
          </strong>
        </div>
      </div>

      {/* Body */}
      <div className="flex flex-col gap-6 p-6">
        {/* Row 1 */}
        <div className="grid grid-cols-1 gap-[18px] md:grid-cols-[300px_1fr] md:gap-[26px]">
          <div>
            <label
              htmlFor="add-hours-date"
              className={labelClass}
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.date')}
            </label>
            <DatePicker
              id="add-hours-date"
              value={date}
              readOnly={locked}
              onChange={(next) => next && onDateChange(next)}
              locale={i18n.language}
              error={!!dateError}
              className={`flow-focus tabular-nums ${fieldSizing}`}
              aria-describedby={dateError ? 'add-hours-date-error' : undefined}
              aria-invalid={dateError ? true : undefined}
              labels={{
                previousMonth: t('common.datePicker.previousMonth'),
                nextMonth: t('common.datePicker.nextMonth'),
              }}
            />
            {dateError && (
              <p
                id="add-hours-date-error"
                className="mt-1 text-meta"
                style={{ color: 'var(--flow-danger-fg)' }}
              >
                {dateError}
              </p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-[18px] md:grid-cols-[96px_96px_1fr]">
            {/* Hours/Minutes are pinned LTR regardless of page direction - a
             * duration reads hours-then-minutes in every language, it is not
             * a layout element that should mirror in RTL. */}
            <div dir="ltr">
              <label
                htmlFor="add-hours-hours"
                className={labelClass}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.add.hours')}
              </label>
              <input
                id="add-hours-hours"
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                readOnly={locked}
                value={hours}
                onChange={(e) => setHours(e.target.value.replace(/\D/g, ''))}
                onFocus={(e) => e.target.select()}
                className={`flow-focus tabular-nums ${fieldSizing}`}
                style={{
                  ...fieldStyle,
                  color: hours === '0' ? 'var(--flow-text-dim)' : fieldStyle.color,
                }}
              />
            </div>
            <div dir="ltr">
              <label
                htmlFor="add-hours-minutes"
                className={labelClass}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.add.minutes')}
              </label>
              <input
                id="add-hours-minutes"
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                readOnly={locked}
                value={minutesRaw}
                onChange={(e) => setMinutesRaw(e.target.value.replace(/\D/g, ''))}
                onFocus={(e) => e.target.select()}
                onBlur={() => setMinutesTouched(true)}
                className={`flow-focus tabular-nums ${fieldSizing}`}
                style={{
                  ...fieldStyle,
                  color: minutesRaw === '00' ? 'var(--flow-text-dim)' : fieldStyle.color,
                  borderColor: minutesError ? 'var(--flow-danger-fg)' : undefined,
                }}
                aria-describedby={minutesError ? 'add-hours-minutes-error' : undefined}
                aria-invalid={minutesError ? true : undefined}
              />
              {minutesError && (
                <p
                  id="add-hours-minutes-error"
                  className="mt-1 text-meta"
                  style={{ color: 'var(--flow-danger-fg)' }}
                >
                  {minutesError}
                </p>
              )}
            </div>

            <div className="col-span-2 md:col-span-1">
              <span className={labelClass} style={{ color: 'var(--flow-text-muted)' }}>
                {t('flow.time.add.location')}
              </span>
              <div
                role="radiogroup"
                aria-label={t('flow.time.add.location')}
                className="grid grid-cols-3 gap-2 md:flex md:flex-wrap"
                onKeyDown={(e) => {
                  if (locked) return;
                  if (e.key !== 'ArrowRight' && e.key !== 'ArrowLeft') return;
                  e.preventDefault();
                  const i = LOCATIONS.indexOf(location);
                  const dir = e.key === 'ArrowRight' ? 1 : -1;
                  setLocation(LOCATIONS[(i + dir + LOCATIONS.length) % LOCATIONS.length]);
                }}
              >
                {LOCATIONS.map((loc) => {
                  const active = loc === location;
                  return (
                    <button
                      key={loc}
                      type="button"
                      role="radio"
                      aria-checked={active}
                      tabIndex={active ? 0 : -1}
                      aria-disabled={locked}
                      onClick={() => !locked && setLocation(loc)}
                      className="flow-focus min-h-[44px] h-auto rounded-xl px-4 py-2 text-body font-semibold whitespace-normal md:min-h-0 md:h-[38px] md:py-0"
                      style={{
                        background: active ? 'var(--accent)' : 'var(--flow-hover)',
                        color: active ? 'var(--flow-on-accent)' : 'var(--flow-text-muted)',
                      }}
                    >
                      {t(`flow.time.mode.${loc}`)}
                    </button>
                  );
                })}
              </div>
            </div>
          </div>
        </div>

        {/* Row 2 */}
        <div className="grid grid-cols-1 gap-[18px] md:grid-cols-[300px_1fr] md:gap-[26px]">
          <div className="flex flex-col gap-[18px]">
            <div>
              <label
                htmlFor="add-hours-project"
                className={labelClass}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.project')}
              </label>
              <select
                id="add-hours-project"
                value={projectId}
                aria-readonly={locked}
                onChange={(e) => {
                  if (locked) return;
                  setProjectId(e.target.value);
                  setTaskId('');
                }}
                className={`flow-focus ${fieldSizing}`}
                style={fieldStyle}
              >
                <option value="">{t('flow.time.add.projectPlaceholder')}</option>
                {(projects ?? []).map((project) => (
                  <option key={project.id} value={project.id}>
                    {project.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label
                htmlFor="add-hours-task"
                className={labelClass}
                style={{ color: 'var(--flow-text-muted)' }}
              >
                {t('flow.time.add.task')}
              </label>
              <select
                id="add-hours-task"
                value={taskId}
                disabled={!projectId}
                aria-readonly={locked}
                title={!projectId ? t('flow.time.add.taskDisabledPlaceholder') : undefined}
                aria-label={
                  !projectId ? t('flow.time.add.taskDisabledPlaceholder') : t('flow.time.add.task')
                }
                onChange={(e) => {
                  if (locked) return;
                  setTaskId(e.target.value);
                }}
                className={`flow-focus ${fieldSizing} disabled:cursor-not-allowed`}
                style={{
                  ...fieldStyle,
                  background: !projectId ? 'var(--flow-hover)' : fieldStyle.background,
                  color: !projectId ? 'var(--flow-text-dim)' : fieldStyle.color,
                  borderColor: !projectId ? undefined : taskId ? 'var(--accent-border)' : undefined,
                }}
              >
                {!projectId ? (
                  <option value="">{t('flow.time.add.taskDisabledPlaceholder')}</option>
                ) : (
                  <>
                    <option value="">{t('flow.time.add.taskPlaceholder')}</option>
                    {(tasks ?? []).map((task) => (
                      <option key={task.id} value={task.id}>
                        {task.identifier ? `${task.identifier} ${task.title}` : task.title}
                      </option>
                    ))}
                  </>
                )}
              </select>
            </div>

            {/* "Your recent tasks" is deliberately not rendered: it would need
             * each of the caller's own last-14-days work days aggregated by
             * task, and there is no endpoint for that (see the "grouped
             * aggregation endpoint" gap in docs/flow-time-tracking.md's Build
             * order). Faking it from only the currently loaded day would not
             * be "last 14 days" at all, so the fallback here is to omit the
             * block and rely on the Task dropdown, per the design spec's own
             * instruction for this gap. The recentTasks/recentTasksHint
             * locale keys stay defined for when that endpoint lands. */}
          </div>

          <div className="flex flex-col gap-[18px]">
            {taskDescription && (
              <div
                className="rounded-lg p-3"
                style={{ background: 'var(--flow-hover)', border: '1px dashed var(--flow-border)' }}
              >
                <div className="flex items-start justify-between gap-2">
                  <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                    {taskDescription}
                  </p>
                  <span
                    className="shrink-0 whitespace-nowrap text-meta font-semibold uppercase tracking-[0.07em]"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    {t('flow.time.add.fromTheTask')}
                  </span>
                </div>
              </div>
            )}

            <div>
              <div className="mb-[7px] flex items-baseline justify-between">
                <label
                  htmlFor="add-hours-note"
                  className="text-meta font-bold uppercase tracking-[0.07em]"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {t('flow.time.description')}
                </label>
                <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.time.add.noteHint')}
                </span>
              </div>
              <textarea
                id="add-hours-note"
                value={note}
                readOnly={locked}
                onChange={(e) => setNote(e.target.value.slice(0, noteMax))}
                className="flow-focus h-[120px] w-full resize-none rounded-lg px-3 py-2 text-title leading-[1.6] md:h-[190px] md:text-meta"
                style={fieldStyle}
              />
              <p
                className="mt-1 text-right text-meta tabular-nums"
                style={{
                  color:
                    note.length >= noteDangerAt
                      ? 'var(--flow-danger-fg)'
                      : 'var(--flow-text-muted)',
                }}
              >
                {t('flow.time.add.noteCounter', { count: note.length, max: noteMax })}
              </p>
            </div>
          </div>
        </div>

        {showsOverCapacityWarning && phase !== 'success' && (
          <div
            className="rounded-lg px-3 py-2 text-meta"
            style={{
              background: 'var(--flow-status-pending-bg)',
              border: '1px solid var(--flow-status-pending-border)',
              color: 'var(--flow-text-muted)',
            }}
          >
            {t('flow.time.add.overCapacity', {
              duration: formatDurationClock(projectedTotalMinutes),
              day: dayLabel,
              target: formatDurationClock(dayTargetMinutes),
            })}
          </div>
        )}

        {phase === 'error' && (
          <div
            ref={bannerRef}
            tabIndex={-1}
            role="alert"
            className="flow-focus rounded-[9px] p-3"
            style={{
              background: 'var(--flow-status-rejected-bg)',
              border: '1px solid var(--flow-status-rejected-border)',
            }}
          >
            <div className="flex items-start gap-2">
              <CircleAlert
                size={15}
                aria-hidden="true"
                style={{ color: 'var(--flow-danger-fg)' }}
                className="mt-0.5 shrink-0"
              />
              <div className="min-w-0 flex-1">
                <p className="text-meta font-semibold" style={{ color: 'var(--flow-danger-fg)' }}>
                  {t('flow.time.add.failedTitle')}
                </p>
                <p className="mt-1 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.time.add.failedBody', { target: targetLabel || '-', ref: errorRef })}
                </p>
                <div className="mt-2 flex gap-2">
                  <button
                    type="button"
                    onClick={() => void handleSubmit()}
                    className="flow-focus h-[30px] rounded-lg px-3 text-meta font-semibold"
                    style={{
                      background: 'transparent',
                      border: '1px solid var(--flow-status-rejected-border)',
                      color: 'var(--flow-danger-fg)',
                    }}
                  >
                    {t('flow.time.add.tryAgain')}
                  </button>
                  <button
                    type="button"
                    onClick={handleCopyNote}
                    className="flow-focus h-[30px] rounded-lg px-3 text-meta font-medium"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {copied ? t('flow.time.add.copyNoteDone') : t('flow.time.add.copyNote')}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {phase === 'success' && successInfo && (
          <div
            ref={bannerRef}
            tabIndex={-1}
            role="status"
            className="flow-focus rounded-[9px] p-3"
            style={{ background: 'var(--accent-muted)', border: '1px solid var(--accent-border)' }}
          >
            <div className="flex items-start gap-2">
              <span
                aria-hidden="true"
                className="mt-1.5 h-[10px] w-[10px] shrink-0 rounded-full"
                style={{ background: 'var(--accent-on-muted)' }}
              />
              <div className="min-w-0 flex-1">
                <p className="text-meta font-semibold" style={{ color: 'var(--accent-on-muted)' }}>
                  {t('flow.time.add.successTitle', {
                    duration: formatDurationClock(successInfo.durationMinutes),
                    target: successInfo.target || '-',
                  })}
                </p>
                <p className="mt-1 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.time.add.successBody', {
                    day: successInfo.dayLabel,
                    total: formatDurationClock(successInfo.newTotalMinutes),
                  })}
                </p>
                <div className="mt-2 flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={resetForLogAnother}
                    className="flow-focus h-[30px] rounded-lg px-3 text-meta font-semibold"
                    style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
                  >
                    {t('flow.time.add.logAnother')}
                  </button>
                  <Link
                    to="/time/reports"
                    className="flow-focus flex h-[30px] items-center rounded-lg px-3 text-meta font-medium"
                    style={{
                      background: 'var(--flow-hover)',
                      border: '1px solid var(--flow-border)',
                      color: 'var(--flow-text)',
                    }}
                  >
                    {t('flow.time.add.seeMyWeek')}
                  </Link>
                  <button
                    type="button"
                    onClick={handleUndo}
                    className="flow-focus h-[30px] rounded-lg px-3 text-meta font-medium"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {t('flow.time.add.undo')}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {locked && (
          <p className="text-meta" style={{ color: 'var(--flow-text-muted)' }} aria-live="polite">
            {t('flow.time.add.submittingLine', {
              duration: formatDurationClock(durationMinutes),
              target: targetLabel || '-',
            })}
          </p>
        )}

        {phase !== 'success' && (
          // Once logged, this "if you submit this, it becomes..." preview
          // would double-count: dayLoggedMinutes already includes the entry
          // just written, and the duration fields are still sitting at what
          // was just submitted (see submitBlocked above), so the two summed
          // together overstate the real total. The success banner already
          // states the real, frozen total - this line only makes sense while
          // something is still pending submission.
          <p aria-live="polite" className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
            <span>{t('flow.time.add.footerLoggingPrefix')}</span>{' '}
            <strong className="tabular-nums font-semibold" style={{ color: 'var(--flow-text)' }}>
              {formatDurationClock(durationMinutes)}
            </strong>{' '}
            <span>{t('flow.time.add.footerMiddle', { day: dayLabel })}</span>{' '}
            <strong className="tabular-nums font-semibold" style={{ color: 'var(--flow-text)' }}>
              {formatDurationClock(projectedTotalMinutes)}
            </strong>
          </p>
        )}
      </div>

      {/* Footer */}
      <div
        className="flex flex-col gap-3 rounded-b-xl px-6 py-[15px] md:flex-row md:items-center md:justify-between"
        style={{ background: 'var(--flow-card)', borderTop: '1px solid var(--flow-border)' }}
      >
        <div className="hidden md:block" />
        <div className="flex gap-2">
          <button
            type="button"
            onClick={handleDiscard}
            disabled={locked}
            className="flow-focus min-h-[44px] flex-1 rounded-lg px-3 text-meta font-semibold disabled:cursor-not-allowed md:flex-none md:h-[38px] md:min-h-0"
            style={{
              background: 'var(--flow-hover)',
              border: '1px solid var(--flow-border)',
              color: locked ? 'var(--flow-text-dim)' : 'var(--flow-text-muted)',
            }}
          >
            {t('flow.time.add.discard')}
          </button>
          <button
            type="button"
            onClick={() => void handleSubmit()}
            disabled={!canSubmit || submitBlocked}
            title={disabledReason || undefined}
            aria-describedby={disabledReason ? 'add-hours-log-reason' : undefined}
            className="flow-focus flex min-h-[44px] flex-1 items-center justify-center gap-1.5 rounded-lg px-4 text-meta font-bold disabled:opacity-40 disabled:cursor-not-allowed md:flex-none md:h-[38px] md:min-h-0"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            {locked ? (
              <>
                <span
                  aria-hidden="true"
                  className="flow-add-hours-spinner inline-block h-3 w-3 rounded-full border-2 border-current"
                  style={{ borderTopColor: 'transparent' }}
                />
                <span style={{ opacity: 0.75 }}>{t('flow.time.add.submittingButton')}</span>
              </>
            ) : durationMinutes > 0 ? (
              <span className="md:hidden">
                {t('flow.time.add.logHoursMobile', {
                  duration: formatDurationClock(durationMinutes),
                })}
              </span>
            ) : null}
            {!locked && (
              <span className={durationMinutes > 0 ? 'hidden md:inline' : ''}>
                {t('flow.time.add.logHours')}
              </span>
            )}
          </button>
        </div>
        {disabledReason && (
          <span id="add-hours-log-reason" className="sr-only">
            {disabledReason}
          </span>
        )}
      </div>
    </div>
  );
};

export default AddHours;
