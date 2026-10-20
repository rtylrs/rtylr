import React, { useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from '@tanstack/react-router';
import {
  CalendarClock,
  Check,
  Mail,
  MessageSquare,
  Phone,
  StickyNote,
  Trash2,
  Users,
} from 'lucide-react';
import { Button } from '../../../shared/components/ui';
import {
  ACTIVITY_TYPES,
  type ActivityType,
  addDaysISO,
  CALL_OUTCOMES,
  type CallOutcome,
  dateOf,
  formatRelative,
  type SalesTask,
  TASK_TYPES,
  type TaskType,
  telHref,
  todayISO,
} from '../data/model';
import {
  useActivities,
  useCompleteTask,
  useCreateTask,
  useDeleteActivity,
  useDeleteTask,
  useLogActivity,
  useTaskList,
} from '../data/queries';
import { DueDate, useNextPageOnScroll, useRepNames } from './common';

export const ACTIVITY_ICONS: Record<string, React.ComponentType<{ size?: number }>> = {
  call: Phone,
  email: Mail,
  meeting: Users,
  sms: MessageSquare,
  note: StickyNote,
};

const FOLLOW_UPS = [
  { key: 'none', days: null },
  { key: 'tomorrow', days: 1 },
  { key: 'inThreeDays', days: 3 },
  { key: 'nextWeek', days: 7 },
  { key: 'pick', days: null },
] as const;

type Subject = { leadId?: string; dealId?: string };

/**
 * Logs a touch on a lead or deal. A call needs its outcome; any touch can
 * schedule the next one as a task.
 */
export const LogActivityPanel: React.FC<{
  subject: Subject;
  phone?: string;
  /** Called after a touch is saved, with the call outcome when it was a call. */
  onLogged?: (outcome?: CallOutcome) => void;
  compact?: boolean;
}> = ({ subject, phone, onLogged, compact }) => {
  const { t } = useTranslation();
  const [type, setType] = useState<ActivityType>('call');
  const [outcome, setOutcome] = useState<CallOutcome | undefined>();
  const [notes, setNotes] = useState('');
  const [minutes, setMinutes] = useState('');
  const [followUp, setFollowUp] = useState<(typeof FOLLOW_UPS)[number]['key']>('none');
  const [followUpDate, setFollowUpDate] = useState(addDaysISO(1));
  const log = useLogActivity();
  const hintId = useId();

  const dueDate =
    followUp === 'none'
      ? undefined
      : followUp === 'pick'
        ? followUpDate
        : addDaysISO(FOLLOW_UPS.find((item) => item.key === followUp)?.days ?? 1);
  const ready = type === 'call' ? Boolean(outcome) : type !== 'note' || notes.trim() !== '';
  const blockedHint = ready
    ? undefined
    : type === 'call'
      ? t('crmSales.activity.pickOutcome')
      : t('crmSales.activity.notesRequired');

  const submit = (): void => {
    if (!ready) return;
    const duration = Number(minutes);
    log.mutate(
      {
        ...subject,
        type,
        outcome: type === 'call' ? outcome : undefined,
        notes: notes.trim() || undefined,
        durationSeconds:
          type === 'call' && Number.isFinite(duration) && duration > 0
            ? Math.round(duration * 60)
            : undefined,
        followUp: dueDate
          ? {
              ...subject,
              type: type === 'note' ? 'todo' : type === 'sms' ? 'call' : type,
              title: t('crmSales.activity.followUpTitle'),
              dueDate,
            }
          : undefined,
      },
      {
        onSuccess: () => {
          const logged = type === 'call' ? outcome : undefined;
          setOutcome(undefined);
          setNotes('');
          setMinutes('');
          setFollowUp('none');
          onLogged?.(logged);
        },
      }
    );
  };

  const href = telHref(phone);
  return (
    <section className="sales-panel" aria-label={t('crmSales.activity.logTitle')}>
      <div className="sales-panel-head">
        <h2>{t('crmSales.activity.logTitle')}</h2>
        {href && !compact && (
          <a className="sales-contact-line" href={href}>
            <Phone size={14} aria-hidden="true" />
            <bdi dir="ltr">{phone}</bdi>
          </a>
        )}
      </div>
      <div className="sales-chip-row" role="group" aria-label={t('crmSales.activity.type')}>
        {ACTIVITY_TYPES.map((item) => {
          const Icon = ACTIVITY_ICONS[item];
          return (
            <button
              key={item}
              type="button"
              className="sales-chip"
              aria-pressed={type === item}
              onClick={() => setType(item)}
            >
              <Icon size={13} />
              {t(`crmSales.activityType.${item}`)}
            </button>
          );
        })}
      </div>
      {type === 'call' && (
        <div className="sales-outcomes" role="group" aria-label={t('crmSales.activity.outcome')}>
          {CALL_OUTCOMES.map((item) => (
            <button
              key={item}
              type="button"
              className="sales-chip"
              aria-pressed={outcome === item}
              onClick={() => setOutcome(item)}
            >
              {t(`crmSales.outcome.${item}`)}
            </button>
          ))}
        </div>
      )}
      <textarea
        className="sales-textarea"
        aria-label={t('crmSales.activity.notes')}
        placeholder={t('crmSales.activity.notesPlaceholder')}
        value={notes}
        onChange={(event) => setNotes(event.target.value)}
      />
      <div className="sales-filters">
        {type === 'call' && (
          <label
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              fontSize: 12,
              color: 'var(--hub-text-muted)',
            }}
          >
            {t('crmSales.activity.minutes')}
            <input
              className="sales-input"
              style={{ width: 70 }}
              inputMode="decimal"
              value={minutes}
              onChange={(event) => setMinutes(event.target.value)}
            />
          </label>
        )}
        <label
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 6,
            fontSize: 12,
            color: 'var(--hub-text-muted)',
          }}
        >
          <CalendarClock size={14} aria-hidden="true" />
          {t('crmSales.activity.followUp')}
          <select
            className="sales-select"
            value={followUp}
            onChange={(event) => setFollowUp(event.target.value as typeof followUp)}
          >
            {FOLLOW_UPS.map((item) => (
              <option key={item.key} value={item.key}>
                {t(`crmSales.activity.followUpOptions.${item.key}`)}
              </option>
            ))}
          </select>
        </label>
        {followUp === 'pick' && (
          <input
            type="date"
            className="sales-input"
            aria-label={t('crmSales.activity.followUpDate')}
            min={todayISO()}
            value={followUpDate}
            onChange={(event) => setFollowUpDate(event.target.value)}
          />
        )}
        <span className="sales-spacer" />
        {blockedHint && (
          <span id={hintId} className="sales-hint">
            {blockedHint}
          </span>
        )}
        <Button
          onClick={submit}
          disabled={!ready || log.isPending}
          aria-describedby={blockedHint ? hintId : undefined}
        >
          {log.isPending ? t('crmSales.common.saving') : t(`crmSales.activity.logAction.${type}`)}
        </Button>
      </div>
    </section>
  );
};

/** A lead's or deal's history, newest first. */
export const Timeline: React.FC<{ subject: Subject }> = ({ subject }) => {
  const { t, i18n } = useTranslation();
  const names = useRepNames();
  const activities = useActivities(subject);
  const sentinel = useNextPageOnScroll(activities);
  const remove = useDeleteActivity();
  const items = activities.data?.pages.flatMap((page) => page.activities ?? []) ?? [];

  return (
    <section className="sales-panel" aria-label={t('crmSales.activity.timeline')}>
      <h2>{t('crmSales.activity.timeline')}</h2>
      {activities.isError ? (
        <p role="alert" className="sales-muted">
          {t('crmSales.activity.loadFailed')}
        </p>
      ) : items.length === 0 ? (
        <p className="sales-muted" style={{ margin: 0, fontSize: 13 }}>
          {activities.isPending ? t('crmSales.common.loading') : t('crmSales.activity.empty')}
        </p>
      ) : (
        <ul className="sales-timeline">
          {items.map((item) => {
            const Icon = ACTIVITY_ICONS[item.type ?? 'note'] ?? StickyNote;
            const title =
              item.type === 'call' && item.outcome
                ? t('crmSales.activity.callWith', {
                    outcome: t(`crmSales.outcome.${item.outcome}`),
                  })
                : t(`crmSales.activityType.${item.type ?? 'note'}`);
            const minutes = item.durationSeconds ? Math.round(item.durationSeconds / 60) : 0;
            return (
              <li key={item.id}>
                <span className="sales-timeline-icon" aria-hidden="true">
                  <Icon size={14} />
                </span>
                <div style={{ minWidth: 0 }}>
                  <div className="sales-timeline-title">
                    {title}
                    {minutes > 0 && (
                      <span className="sales-muted" style={{ fontWeight: 400 }}>
                        {' · '}
                        {t('crmSales.activity.minutesShort', { count: minutes })}
                      </span>
                    )}
                  </div>
                  {item.notes && <p className="sales-timeline-notes">{item.notes}</p>}
                  <div className="sales-cell-sub">{names.get(item.createdBy ?? '') ?? ''}</div>
                </div>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 4 }}>
                  <time
                    className="sales-timeline-time"
                    dateTime={dateOf(item.occurredAt)?.toISOString()}
                  >
                    {formatRelative(dateOf(item.occurredAt), i18n.language)}
                  </time>
                  <button
                    type="button"
                    className="sales-chip"
                    style={{ minHeight: 24, padding: 4 }}
                    aria-label={t('crmSales.activity.delete')}
                    onClick={() => item.id && remove.mutate(item.id)}
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              </li>
            );
          })}
        </ul>
      )}
      {activities.hasNextPage && (
        <Button
          ref={sentinel}
          variant="outline"
          onClick={() => void activities.fetchNextPage()}
          disabled={activities.isFetchingNextPage}
        >
          {activities.isFetchingNextPage
            ? t('crmSales.common.loading')
            : t('crmSales.common.loadMore')}
        </Button>
      )}
    </section>
  );
};

export const TaskRow: React.FC<{
  task: SalesTask;
  showSubject?: boolean;
  repName?: string;
}> = ({ task, showSubject, repName }) => {
  const { t } = useTranslation();
  const complete = useCompleteTask();
  const remove = useDeleteTask();
  const done = Boolean(task.completedAt);
  const href = task.type === 'call' ? telHref(task.phone) : undefined;
  const subject = task.dealTitle || task.companyName || task.leadName;
  return (
    <div className={`sales-task${done ? ' is-done' : ''}`}>
      <input
        type="checkbox"
        checked={done}
        aria-label={done ? t('crmSales.tasks.reopen') : t('crmSales.tasks.complete')}
        onChange={() => task.id && complete.mutate({ id: task.id, done: !done })}
      />
      <div className="sales-task-main">
        <div className="sales-task-title">
          {task.title}{' '}
          <span className="sales-badge">{t(`crmSales.taskType.${task.type ?? 'todo'}`)}</span>
        </div>
        <div className="sales-cell-sub">
          {showSubject && subject && (
            <>
              {task.dealId ? (
                <Link className="sales-link" to={`/crm/sales/deals/${task.dealId}`}>
                  {subject}
                </Link>
              ) : task.leadId ? (
                <Link className="sales-link" to={`/crm/sales/leads/${task.leadId}`}>
                  {subject}
                </Link>
              ) : (
                subject
              )}
              {' · '}
            </>
          )}
          {href && (
            <>
              <a className="sales-link" href={href}>
                <bdi dir="ltr">{task.phone}</bdi>
              </a>
              {' · '}
            </>
          )}
          {repName}
        </div>
      </div>
      <DueDate date={task.dueDate} />
      <button
        type="button"
        className="sales-chip"
        style={{ minHeight: 26, padding: 4 }}
        aria-label={t('crmSales.tasks.delete')}
        onClick={() => task.id && remove.mutate(task.id)}
      >
        <Trash2 size={12} />
      </button>
    </div>
  );
};

/** Open tasks for a lead or deal, with a quick way to add one. */
export const SubjectTasks: React.FC<{ subject: Subject }> = ({ subject }) => {
  const { t } = useTranslation();
  const names = useRepNames();
  const tasks = useTaskList({ scope: 'open', rep: 'all', due: '', ...subject }, 20);
  const create = useCreateTask();
  const [title, setTitle] = useState('');
  const [type, setType] = useState<TaskType>('call');
  const [dueDate, setDueDate] = useState(addDaysISO(1));
  const items = tasks.data?.pages.flatMap((page) => page.tasks ?? []) ?? [];

  const add = (event: React.FormEvent): void => {
    event.preventDefault();
    create.mutate(
      { ...subject, title: title.trim() || t(`crmSales.taskType.${type}`), type, dueDate },
      { onSuccess: () => setTitle('') }
    );
  };

  return (
    <section className="sales-panel" aria-label={t('crmSales.tasks.title')}>
      <h2>{t('crmSales.tasks.title')}</h2>
      {items.length === 0 ? (
        <p className="sales-muted" style={{ margin: 0, fontSize: 13 }}>
          {tasks.isPending ? t('crmSales.common.loading') : t('crmSales.tasks.noneOpen')}
        </p>
      ) : (
        <div>
          {items.map((task) => (
            <TaskRow key={task.id} task={task} repName={names.get(task.assignedTo ?? '')} />
          ))}
        </div>
      )}
      <form onSubmit={add} className="sales-filters">
        <input
          className="sales-input"
          style={{ flex: '1 1 140px' }}
          aria-label={t('crmSales.tasks.newTitle')}
          placeholder={t('crmSales.tasks.newTitle')}
          value={title}
          onChange={(event) => setTitle(event.target.value)}
        />
        <select
          className="sales-select"
          aria-label={t('crmSales.tasks.type')}
          value={type}
          onChange={(event) => setType(event.target.value as TaskType)}
        >
          {TASK_TYPES.map((item) => (
            <option key={item} value={item}>
              {t(`crmSales.taskType.${item}`)}
            </option>
          ))}
        </select>
        <input
          type="date"
          className="sales-input"
          aria-label={t('crmSales.tasks.dueDate')}
          value={dueDate}
          onChange={(event) => setDueDate(event.target.value)}
        />
        <Button type="submit" variant="outline" disabled={create.isPending || !dueDate}>
          <Check size={13} />
          {t('crmSales.tasks.add')}
        </Button>
      </form>
    </section>
  );
};
