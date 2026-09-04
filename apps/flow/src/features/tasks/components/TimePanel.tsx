import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Timer, Trash2, RotateCcw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Avatar, EmptyState } from '@voxire/ui';
import { useTimeLogs, useLogTime, useDeleteTimeLog } from '../useTimeLogs';
import { useFlowTeam } from '../../../shared/hooks/queries/useFlowTeam';
import { initialsFromName } from '../../../shared/lib/memberUtils';
import { formatEntryTimestamp, todayIso } from '../../../shared/lib/dateUtils';
import { formatDurationClock, parseDurationInput } from '../../../shared/lib/durationUtils';
import GateNotice from '../../../shared/components/GateNotice';
import DurationField from './DurationField';
import { roundLoggedMinutes, useTimeSettingsStore } from '../../../shared/stores/timeSettingsStore';
import { useRealtimeStatus } from '../../../shared/lib/realtime';

interface TimePanelProps {
  taskId: string;
  /** Task estimate in hours, or null when none is set. */
  estimateHours: number | null;
  /** Logging your own time is the universal floor, so this is false only for a viewer. */
  canLog: boolean;
  /** Logging against another person needs "manage team time". */
  canLogForOthers: boolean;
  currentUserId: string | null;
}

/**
 * The time block in the task-detail rail.
 *
 * Replaces the inline section that lived in TaskDetail.tsx. Reads and writes
 * through the three time-log RPCs that exist today; anything the contract has
 * no field for is absent rather than drawn inert. See
 * docs/flow-time-tracking.md for what that excludes and why.
 *
 * Below 768px this is not a rail: it stacks under the description, so the base
 * styles are the phone ones and `md:` tightens them.
 */
const TimePanel = ({
  taskId,
  estimateHours,
  canLog,
  canLogForOthers,
  currentUserId,
}: TimePanelProps): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const { data, isLoading, isError, refetch } = useTimeLogs(taskId);
  const logTime = useLogTime();
  const deleteTimeLog = useDeleteTimeLog();
  const { data: team } = useFlowTeam();
  const roundingMinutes = useTimeSettingsStore((s) => s.roundingMinutes);
  const realtimeStatus = useRealtimeStatus();

  const [formOpen, setFormOpen] = useState(false);
  const [duration, setDuration] = useState('');
  const [description, setDescription] = useState('');
  // Wraps whichever trigger is currently mounted - the plain "log time"
  // button, or the shared EmptyState's own action button when the list is
  // empty. The shared component exposes no ref of its own, so focus is
  // restored by querying the one button inside this wrapper instead of
  // holding a ref to a specific trigger element.
  const triggerContainerRef = useRef<HTMLDivElement | null>(null);

  const logs = data?.timeLogs ?? [];
  const totalMinutes = logs.reduce((sum, l) => sum + l.durationMinutes, 0);
  const estimateMinutes = estimateHours != null ? Math.round(estimateHours * 60) : null;

  const nameFor = useMemo(() => {
    const byId = new Map((team ?? []).map((m) => [m.id, m.name]));
    return (userId: number): string => byId.get(String(userId)) ?? t('flow.time.unknownAuthor');
  }, [team, t]);

  // The mutation's own in-flight variables are the optimistic row. Holding a
  // duplicate copy in local state would leave a phantom entry behind whenever
  // the mutation settled through a path this component did not observe.
  const pendingMinutes =
    logTime.isPending && logTime.variables ? logTime.variables.duration_minutes : null;
  const pendingDescription =
    logTime.isPending && logTime.variables ? logTime.variables.description : undefined;
  // Drives both the empty-state illustration below and whether the regular
  // "log time" trigger also renders - showing both would just be the same
  // action offered twice.
  const isTrulyEmpty = !isLoading && !isError && logs.length === 0 && pendingMinutes === null;

  // Focus cannot be restored inside closeForm: the trigger is unmounted while
  // the form is open, so the ref still points at the detached node and the
  // focus call lands nowhere. Restoring in an effect runs after the trigger is
  // back in the tree.
  const restoreFocus = useRef(false);
  useEffect(() => {
    if (!formOpen && restoreFocus.current) {
      restoreFocus.current = false;
      triggerContainerRef.current?.querySelector('button')?.focus();
    }
  }, [formOpen]);

  const closeForm = (): void => {
    restoreFocus.current = true;
    setFormOpen(false);
    setDuration('');
    setDescription('');
  };

  const submit = (): void => {
    const minutes = parseDurationInput(duration);
    if (minutes === null) return;
    logTime.mutate(
      {
        task_id: Number(taskId),
        duration_minutes: roundLoggedMinutes(minutes, roundingMinutes),
        description: description.trim() || undefined,
      },
      {
        onSuccess: () => closeForm(),
        // On failure the form stays open with the values still in it. An entry
        // that vanished silently is the one outcome worth guarding against.
        onError: () => undefined,
      }
    );
  };

  // What will actually be stored, previewed on the submit button. Rounding that
  // the person cannot see before they commit is rounding they will dispute later.
  const parsedPreview = parseDurationInput(duration);
  const roundedPreview =
    parsedPreview === null ? null : roundLoggedMinutes(parsedPreview, roundingMinutes);
  const wasRounded = parsedPreview !== null && roundedPreview !== parsedPreview;

  return (
    <section className="mb-5" aria-labelledby={`time-panel-${taskId}`}>
      <div className="mb-1.5 flex items-center gap-1.5">
        <Timer size={11} aria-hidden="true" />
        <h3
          id={`time-panel-${taskId}`}
          className="text-meta font-semibold uppercase tracking-wider"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          {t('flow.time.heading')}
        </h3>
      </div>

      {/* Running total. aria-live so a screen reader hears the new figure after
       * a log, without the whole panel being re-announced. */}
      <div aria-live="polite" aria-atomic="true" className="flex items-baseline gap-1.5">
        <span className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
          {formatDurationClock(totalMinutes)}
        </span>
        {estimateMinutes !== null && (
          <>
            <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.time.of')}
            </span>
            <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {formatDurationClock(estimateMinutes)}
            </span>
          </>
        )}
      </div>

      {estimateMinutes !== null && totalMinutes > 0 && (
        <EstimateBar
          loggedMinutes={totalMinutes}
          estimateMinutes={estimateMinutes}
          label={t('flow.time.estimateBarLabel')}
          overLabel={t('flow.time.overEstimate')}
          percentLabel={(percent) => t('flow.time.percentOfEstimate', { percent })}
        />
      )}

      {/* The one place in this feature where silence risks losing work: a stale
       * list looks identical to an empty one. Writes are unaffected, and the copy
       * says so rather than implying the panel is broken. */}
      {(realtimeStatus === 'reconnecting' || realtimeStatus === 'disconnected') && (
        <p
          role="status"
          className="mt-2 rounded-sm px-2 py-1 text-meta leading-snug"
          style={{
            background: 'var(--flow-status-pending-bg)',
            border: '1px solid var(--flow-status-pending-border)',
            color: 'var(--flow-status-pending)',
          }}
        >
          {t(
            realtimeStatus === 'reconnecting'
              ? 'flow.time.socketReconnecting'
              : 'flow.time.socketDisconnected'
          )}
        </p>
      )}

      {isLoading && (
        <div className="mt-2 space-y-1">
          <EntrySkeleton />
          <EntrySkeleton />
        </div>
      )}

      {isError && (
        <div
          className="mt-2 rounded-md p-2"
          style={{
            background: 'var(--flow-status-rejected-bg)',
            border: '1px solid var(--flow-status-rejected-border)',
          }}
        >
          <p className="text-meta leading-snug" style={{ color: 'var(--flow-text)' }}>
            {t('flow.time.loadFailed')}
          </p>
          <button
            type="button"
            onClick={() => void refetch()}
            className="flow-focus mt-1.5 inline-flex min-h-[44px] items-center gap-1 rounded-sm px-2 text-meta font-semibold md:min-h-0 md:py-1"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text)' }}
          >
            <RotateCcw size={11} aria-hidden="true" />
            {t('flow.time.retry')}
          </button>
        </div>
      )}

      {isTrulyEmpty && (
        <div className="mt-2" ref={triggerContainerRef}>
          <EmptyState
            icon={Timer}
            title={t('flow.time.empty')}
            action={
              canLog
                ? { label: t('flow.time.logFirst'), onClick: () => setFormOpen(true) }
                : { label: t('flow.time.retry'), onClick: () => void refetch() }
            }
            size="compact"
          />
        </div>
      )}

      {(logs.length > 0 || pendingMinutes !== null) && (
        <ul className="mt-2 space-y-1">
          {logs.map((log) => {
            const authorName = nameFor(log.userId);
            // Deleting is gated on the same permission as logging, not on
            // ownership alone: a view-only user who could delete but not add
            // would be a strictly stranger permission than either.
            const canDelete =
              canLog && currentUserId != null && String(log.userId) === currentUserId;
            // flowApi substitutes '-' when a description is omitted, because
            // the contract marks it @required. Do not render the placeholder.
            const body = log.description === '-' ? '' : log.description;
            return (
              <li
                key={log.id}
                className="flex items-start gap-2 rounded-sm px-2 py-2 md:py-1.5"
                style={{ background: 'var(--flow-hover)' }}
              >
                <Avatar size="sm" initials={initialsFromName(authorName)} aria-hidden="true" />
                <div className="min-w-0 flex-1">
                  <span className="text-meta font-semibold" style={{ color: 'var(--accent-fg)' }}>
                    {formatDurationClock(log.durationMinutes)}
                  </span>
                  {body && (
                    <p
                      className="mt-0.5 text-meta leading-snug"
                      style={{ color: 'var(--flow-text)' }}
                    >
                      {body}
                    </p>
                  )}
                  <p className="mt-0.5 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                    <span>{authorName}</span>
                    {' · '}
                    {formatEntryTimestamp(log.loggedAt || log.createdAt, i18n.language)}
                  </p>
                </div>
                {canDelete && (
                  <button
                    type="button"
                    onClick={() => deleteTimeLog.mutate({ id: log.id, taskId })}
                    aria-label={t('flow.time.deleteEntryLabel', {
                      duration: formatDurationClock(log.durationMinutes),
                    })}
                    className="flow-focus flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm md:h-6 md:w-6"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <Trash2 size={12} aria-hidden="true" />
                  </button>
                )}
              </li>
            );
          })}

          {pendingMinutes !== null && (
            <li
              className="flex items-start gap-2 rounded-sm px-2 py-2 opacity-60 md:py-1.5"
              style={{ background: 'var(--flow-hover)' }}
            >
              <div className="min-w-0 flex-1">
                <span className="text-meta font-semibold" style={{ color: 'var(--accent-fg)' }}>
                  {formatDurationClock(pendingMinutes)}
                </span>
                {pendingDescription && pendingDescription !== '-' && (
                  <p
                    className="mt-0.5 text-meta leading-snug"
                    style={{ color: 'var(--flow-text)' }}
                  >
                    {pendingDescription}
                  </p>
                )}
                <p className="mt-0.5 text-meta" style={{ color: 'var(--flow-text-muted)' }}>
                  {t('flow.time.saving')}
                </p>
              </div>
            </li>
          )}
        </ul>
      )}

      {canLog && !formOpen && !isTrulyEmpty && (
        <div ref={triggerContainerRef}>
          <button
            type="button"
            onClick={() => setFormOpen(true)}
            className="flow-focus mt-2 inline-flex min-h-[44px] w-full items-center justify-center rounded-sm px-3 text-meta font-semibold md:min-h-0 md:w-auto md:py-1.5"
            style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
          >
            {t('flow.time.logTime')}
          </button>
        </div>
      )}

      {canLog && formOpen && (
        <div className="mt-2 space-y-2">
          <DurationField
            id={`time-duration-${taskId}`}
            label={t('flow.time.duration')}
            value={duration}
            onChange={setDuration}
            onSubmit={submit}
            onCancel={closeForm}
            placeholder={t('flow.time.durationPlaceholder')}
            hint={t('flow.time.durationHint')}
            invalidMessage={t('flow.time.durationInvalid')}
            autoFocus
          />

          {/* The create RPC has no date field, so this is fixed to today and
           * says so. Root CLAUDE.md rule 12: a disabled control states its
           * reason or is absent, and here the reason is the point. */}
          <div>
            <label
              htmlFor={`time-date-${taskId}`}
              className="mb-1 block text-meta font-semibold uppercase tracking-wider"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.date')}
            </label>
            <input
              id={`time-date-${taskId}`}
              type="date"
              value={todayIso()}
              readOnly
              disabled
              aria-describedby={`time-date-gate-${taskId}`}
              className="min-h-[44px] w-full rounded-sm px-2 py-2 text-title md:min-h-0 md:py-1 md:text-meta"
              style={{
                background: 'var(--flow-hover)',
                border: '1px solid var(--flow-border)',
                color: 'var(--flow-text-muted)',
              }}
            />
            <GateNotice
              id={`time-date-gate-${taskId}`}
              className="mt-1"
              reason={t('flow.time.dateFixedToToday')}
            />
          </div>

          <div>
            <label
              htmlFor={`time-description-${taskId}`}
              className="mb-1 block text-meta font-semibold uppercase tracking-wider"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.description')}
            </label>
            <input
              id={`time-description-${taskId}`}
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  e.preventDefault();
                  submit();
                }
                if (e.key === 'Escape') {
                  e.preventDefault();
                  closeForm();
                }
              }}
              className="flow-focus min-h-[44px] w-full rounded-sm px-2 py-2 text-title md:min-h-0 md:py-1 md:text-meta"
              style={{
                background: 'var(--flow-input-bg)',
                border: '1px solid var(--flow-border)',
                color: 'var(--flow-text)',
              }}
            />
          </div>

          {wasRounded && (
            <p className="text-meta" style={{ color: 'var(--flow-status-pending)' }}>
              {/* `count` rather than a plain number: i18next selects the plural
               * form from it, and Arabic needs six of them. */}
              {t('flow.time.roundedNotice', {
                count: roundingMinutes,
                amount: formatDurationClock(roundedPreview as number),
              })}
            </p>
          )}

          {!canLogForOthers && <GateNotice reason={t('flow.time.logForOthersGate')} />}

          <div className="flex gap-1.5">
            <button
              type="button"
              onClick={submit}
              disabled={parsedPreview === null || logTime.isPending}
              className="flow-focus min-h-[44px] flex-1 rounded-sm px-3 text-meta font-semibold disabled:opacity-40 md:min-h-0 md:py-1.5"
              style={{ background: 'var(--accent)', color: 'var(--flow-on-accent)' }}
            >
              {roundedPreview !== null
                ? t('flow.time.logAmount', { amount: formatDurationClock(roundedPreview) })
                : t('flow.time.logAction')}
            </button>
            <button
              type="button"
              onClick={closeForm}
              className="flow-focus min-h-[44px] rounded-sm px-3 text-meta md:min-h-0 md:py-1.5"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
            >
              {t('flow.time.cancel')}
            </button>
          </div>
        </div>
      )}
    </section>
  );
};

const EntrySkeleton = (): React.JSX.Element => (
  <div
    className="flow-shimmer h-10 rounded-sm md:h-8"
    style={{
      background:
        'linear-gradient(90deg, var(--flow-hover) 25%, var(--flow-active) 50%, var(--flow-hover) 75%)',
      backgroundSize: '200% 100%',
    }}
  />
);

interface EstimateBarProps {
  loggedMinutes: number;
  estimateMinutes: number;
  label: string;
  overLabel: string;
  percentLabel: (percent: number) => string;
}

/**
 * Logged against estimate. The fill is capped at 100% width so a 300% overrun
 * does not blow out the rail; the overrun is carried by the tone and the pill
 * instead. One pill in the warn tone, not a red card - the rail is not the
 * place to litigate an overrun, it only has to be visible.
 */
const EstimateBar = ({
  loggedMinutes,
  estimateMinutes,
  label,
  overLabel,
  percentLabel,
}: EstimateBarProps): React.JSX.Element => {
  const percent = Math.round((loggedMinutes / estimateMinutes) * 100);
  const over = loggedMinutes > estimateMinutes;
  const tone = over ? 'var(--flow-capacity-over)' : 'var(--flow-capacity-at)';

  return (
    <div className="mt-1.5">
      <div
        role="progressbar"
        aria-label={label}
        aria-valuenow={percent}
        aria-valuemin={0}
        aria-valuetext={percentLabel(percent)}
        className="h-1 w-full overflow-hidden rounded-full"
        style={{ background: 'var(--flow-active)' }}
      >
        <div
          className="h-full rounded-full"
          style={{ width: `${Math.min(percent, 100)}%`, background: tone }}
        />
      </div>
      <div className="mt-1 flex items-center gap-1.5">
        <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {percentLabel(percent)}
        </span>
        {over && (
          <span
            className="rounded-full px-2 text-meta font-semibold"
            style={{
              background: 'var(--flow-status-pending-bg)',
              border: '1px solid var(--flow-status-pending-border)',
              color: 'var(--flow-status-pending)',
            }}
          >
            {overLabel}
          </span>
        )}
      </div>
    </div>
  );
};

export default TimePanel;
