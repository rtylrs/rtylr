import React, { useId, useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { TFunction } from 'i18next';
import {
  Activity,
  Calendar as CalendarIcon,
  CircleDot,
  Flag,
  Tag,
  Timer,
  Type,
  User as UserIcon,
} from 'lucide-react';
import { EmptyState } from '../../../../../shared/components/dash/EmptyState';
import type { ApiTaskActivity, ApiTimeLog } from '../../../shared/services/flowApi';
import type { Comment } from '../../../shared/stores/flowStore';
import { formatDurationClock } from '../../../shared/lib/durationUtils';
import { formatEntryTimestamp } from '../../../shared/lib/dateUtils';
import { hasQueryFailed } from '../../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../../shared/components/ListLoadFailed';
import { useTimeLogs } from '../useTimeLogs';

export type ActivityTab = 'all' | 'comments' | 'history' | 'time';
export const ACTIVITY_TABS: readonly ActivityTab[] = ['all', 'comments', 'history', 'time'];
const DEFAULT_TAB: ActivityTab = 'comments';

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatRelative(dateStr: string, t: TFunction): string {
  const d = new Date(dateStr);
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return t('flow.taskDetail.justNow');
  if (mins < 60) return t('flow.taskDetail.minutesAgo', { count: mins });
  const hours = Math.floor(mins / 60);
  if (hours < 24) return t('flow.taskDetail.hoursAgo', { count: hours });
  const days = Math.floor(hours / 24);
  if (days < 30) return t('flow.taskDetail.daysAgo', { count: days });
  return d.toLocaleDateString(t('flow.taskDetail.dateLocale'), {
    month: 'short',
    day: 'numeric',
  });
}

function humanizeValue(action: string, value: string): string {
  if (!value) return '-';
  if (action.includes('status') || action === 'status_changed' || action === 'status_change') {
    return value
      .split('_')
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(' ');
  }
  if (
    action.includes('priority') ||
    action === 'priority_changed' ||
    action === 'priority_change'
  ) {
    return value.charAt(0).toUpperCase() + value.slice(1);
  }
  return value;
}

export function describeActivity(
  a: ApiTaskActivity,
  t: TFunction,
  peopleNames: Record<string, string>
): React.ReactNode {
  const action = a.action;
  const oldVal = humanizeValue(action, a.oldValue);
  const newVal = humanizeValue(action, a.newValue);

  if (action === 'link_added') {
    return t('flow.taskDetail.linkAdded', { value: a.newValue });
  }
  if (action === 'link_removed') {
    return t('flow.taskDetail.linkRemoved', { value: a.oldValue });
  }

  if (action === 'status_changed' || action === 'status_change') {
    return t('flow.taskDetail.statusChanged', { from: oldVal, to: newVal });
  }
  if (action === 'priority_changed' || action === 'priority_change') {
    return t('flow.taskDetail.priorityChanged', { from: oldVal, to: newVal });
  }
  if (action === 'assignee_changed' || action === 'assignee_change') {
    // The server records user ids here, and 0 for unassigned, so the row read
    // "assigned to 4217" until the id was resolved to a name.
    const to =
      !a.newValue || a.newValue === '0'
        ? t('flow.quickCreate.unassigned')
        : (peopleNames[a.newValue] ?? t('flow.time.unknownAuthor'));
    return t('flow.taskDetail.assigneeChanged', { to });
  }
  if (action === 'title_changed' || action === 'title_change') {
    return t('flow.taskDetail.titleChanged');
  }
  if (action === 'due_date_changed' || action === 'due_date_change') {
    return t('flow.taskDetail.dueDateChanged', { to: newVal });
  }
  // generic
  const label = action
    .replace(/_/g, ' ')
    .replace(/\bchanged\b/, 'changed')
    .trim();
  return a.newValue
    ? t('flow.taskDetail.genericActivity', { label, value: newVal })
    : t('flow.taskDetail.genericActivityNoValue', { label });
}

function activityIcon(action: string): React.ComponentType<{ size?: number }> {
  if (action.includes('status')) return CircleDot;
  if (action.includes('priority')) return Flag;
  if (action.includes('assignee')) return UserIcon;
  if (action.includes('title')) return Type;
  if (action.includes('due_date')) return CalendarIcon;
  if (action.includes('label')) return Tag;
  return Activity;
}

/**
 * History is the task's change log without its comments. The server writes a
 * `comment_added` row for every comment, carrying the comment's text, so
 * leaving them in repeated each comment in the log it was meant to skip.
 */
export function historyEntries(entries: readonly ApiTaskActivity[]): ApiTaskActivity[] {
  return entries.filter((entry) => entry.action !== 'comment_added');
}

export type FeedItem =
  | { kind: 'comment'; key: string; at: number; comment: Comment }
  | { kind: 'history'; key: string; at: number; entry: ApiTaskActivity }
  | { kind: 'time'; key: string; at: number; log: ApiTimeLog };

function timeOf(iso: string | undefined): number {
  const at = iso ? Date.parse(iso) : Number.NaN;
  return Number.isNaN(at) ? 0 : at;
}

/**
 * Everything that happened on the task, oldest first, so it reads top to
 * bottom like the comment thread and ends at the comment box. A time entry is
 * placed by when it was recorded, not the day it was logged against: a
 * Friday entry made on Monday belongs after Monday's comments.
 */
export function buildFeed(
  comments: readonly Comment[],
  activity: readonly ApiTaskActivity[],
  timeLogs: readonly ApiTimeLog[]
): FeedItem[] {
  const items: FeedItem[] = [
    ...comments.map((comment) => ({
      kind: 'comment' as const,
      key: `comment-${comment.id}`,
      at: timeOf(comment.createdAt),
      comment,
    })),
    ...historyEntries(activity).map((entry) => ({
      kind: 'history' as const,
      key: `history-${entry.id}`,
      at: timeOf(entry.createdAt),
      entry,
    })),
    ...timeLogs.map((log) => ({
      kind: 'time' as const,
      key: `time-${log.id}`,
      at: timeOf(log.createdAt || log.loggedAt),
      log,
    })),
  ];
  // Array.prototype.sort is stable, so equal times keep the order above.
  return items.sort((a, b) => a.at - b.at);
}

// The last tab a person chose, per project. Kept in this browser only: there
// is no server-side preference for it, and losing it just means the default.
function tabStorageKey(userId: string | undefined, projectId: string | undefined): string {
  return `flow.taskActivityTab.${userId ?? 'anon'}.${projectId ?? 'none'}`;
}

function readStoredTab(key: string): ActivityTab {
  try {
    const stored = window.localStorage.getItem(key);
    if (stored && (ACTIVITY_TABS as readonly string[]).includes(stored)) {
      return stored as ActivityTab;
    }
  } catch {
    // Storage can be blocked (private mode, policy). The default is fine.
  }
  return DEFAULT_TAB;
}

function storeTab(key: string, tab: ActivityTab): void {
  try {
    window.localStorage.setItem(key, tab);
  } catch {
    // See readStoredTab.
  }
}

// ── Rows ─────────────────────────────────────────────────────────────────────

const RowIcon: React.FC<{ icon: React.ComponentType<{ size?: number }> }> = ({ icon: Icon }) => (
  <span
    aria-hidden="true"
    className="h-5 w-5 rounded-full flex items-center justify-center shrink-0 mt-0.5"
    style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
  >
    <Icon size={10} />
  </span>
);

const HistoryRow: React.FC<{
  entry: ApiTaskActivity;
  peopleNames: Record<string, string>;
  t: TFunction;
}> = ({ entry, peopleNames, t }) => (
  <li className="flex items-start gap-2.5">
    <RowIcon icon={activityIcon(entry.action)} />
    <div className="flex-1 min-w-0 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
      {describeActivity(entry, t, peopleNames)}
      <span className="ms-2" style={{ color: 'var(--flow-text-dim)' }}>
        {formatRelative(entry.createdAt, t)}
      </span>
    </div>
  </li>
);

const TimeRow: React.FC<{
  log: ApiTimeLog;
  peopleNames: Record<string, string>;
  t: TFunction;
  locale: string;
}> = ({ log, peopleNames, t, locale }) => {
  const name = peopleNames[String(log.userId)] ?? t('flow.time.unknownAuthor');
  // flowApi substitutes '-' for an omitted description; it is not text.
  const body = log.description === '-' ? '' : log.description;
  return (
    <li className="flex items-start gap-2.5">
      <RowIcon icon={Timer} />
      <div className="flex-1 min-w-0 text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.taskDetail.activityTabs.timeLogged', {
          name,
          duration: formatDurationClock(log.durationMinutes),
        })}
        <span className="ms-2" style={{ color: 'var(--flow-text-dim)' }}>
          {formatEntryTimestamp(log.loggedAt || log.createdAt, locale)}
        </span>
        {body && (
          <p
            className="mt-0.5 leading-snug whitespace-pre-wrap break-words"
            style={{ color: 'var(--flow-text)' }}
          >
            {body}
          </p>
        )}
      </div>
    </li>
  );
};

// ── Tabs ─────────────────────────────────────────────────────────────────────

interface TaskActivityTabsProps {
  taskId: string;
  projectId?: string;
  currentUserId?: string;
  comments: Comment[];
  activity: ApiTaskActivity[];
  peopleNames: Record<string, string>;
  /** Renders comments with the panel's own edit and mention wiring. */
  renderComments: (comments: Comment[]) => React.ReactNode;
  /** The comment box, shown under All and Comments. */
  composer: React.ReactNode;
}

const TAB_LABEL_KEYS: Record<ActivityTab, string> = {
  all: 'flow.taskDetail.activityTabs.all',
  comments: 'flow.taskDetail.comments',
  history: 'flow.taskDetail.activityTabs.history',
  time: 'flow.taskDetail.activityTabs.time',
};

/**
 * The task's conversation and history, split the way Jira splits it: All,
 * Comments, History and Time. Time is the task's existing time entries, read
 * from the same query the time panel uses, so nothing extra is fetched.
 */
const TaskActivityTabs: React.FC<TaskActivityTabsProps> = ({
  taskId,
  projectId,
  currentUserId,
  comments,
  activity,
  peopleNames,
  renderComments,
  composer,
}) => {
  const { t, i18n } = useTranslation();
  const baseId = useId();
  const storageKey = tabStorageKey(currentUserId, projectId);
  const [choice, setChoice] = useState(() => ({
    key: storageKey,
    tab: readStoredTab(storageKey),
  }));
  // The drawer can swap to a task in another project without remounting.
  const tab = choice.key === storageKey ? choice.tab : readStoredTab(storageKey);
  const selectTab = (next: ActivityTab): void => {
    setChoice({ key: storageKey, tab: next });
    storeTab(storageKey, next);
  };

  const timeLogsQuery = useTimeLogs(taskId);
  const timeLogs = timeLogsQuery.data?.timeLogs ?? [];
  const timeFailed = hasQueryFailed(timeLogsQuery);
  const history = historyEntries(activity);

  const tabId = (name: ActivityTab): string => `${baseId}-tab-${name}`;
  const panelId = `${baseId}-panel`;

  // Arrow keys move along the tabs in reading order, so they flip in RTL.
  const onTabKeyDown = (event: React.KeyboardEvent<HTMLDivElement>): void => {
    // From the focused tab: focus can sit on a tab that is not selected yet.
    const focused = ACTIVITY_TABS.findIndex(
      (name) => document.getElementById(tabId(name)) === document.activeElement
    );
    const at = focused >= 0 ? focused : ACTIVITY_TABS.indexOf(tab);
    const rtl = getComputedStyle(event.currentTarget).direction === 'rtl';
    let next: number | null = null;
    if (event.key === 'ArrowRight') next = rtl ? at - 1 : at + 1;
    if (event.key === 'ArrowLeft') next = rtl ? at + 1 : at - 1;
    if (event.key === 'Home') next = 0;
    if (event.key === 'End') next = ACTIVITY_TABS.length - 1;
    if (next === null) return;
    event.preventDefault();
    const target = ACTIVITY_TABS[(next + ACTIVITY_TABS.length) % ACTIVITY_TABS.length];
    selectTab(target);
    document.getElementById(tabId(target))?.focus();
  };

  const counts: Partial<Record<ActivityTab, number>> = {
    comments: comments.length,
    time: timeLogs.length,
  };

  const renderPanel = (): React.ReactNode => {
    if (tab === 'comments') {
      return (
        <>
          <div className="mb-4">{renderComments(comments)}</div>
          {composer}
        </>
      );
    }
    if (tab === 'history') {
      if (history.length === 0) {
        return (
          <EmptyState
            icon={Activity}
            title={t('flow.taskDetail.noActivity')}
            action={{
              label: t('flow.taskDetail.goToComments'),
              onClick: () => selectTab('comments'),
            }}
            size="compact"
          />
        );
      }
      return (
        <ul className="space-y-3">
          {history.map((entry) => (
            <HistoryRow key={entry.id} entry={entry} peopleNames={peopleNames} t={t} />
          ))}
        </ul>
      );
    }
    if (tab === 'time') {
      if (timeFailed) return <ListLoadFailed onRetry={() => void timeLogsQuery.refetch()} />;
      if (timeLogs.length === 0) {
        return timeLogsQuery.isPending ? null : (
          <p className="text-flow-meta py-6 text-center" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.time.empty')}
          </p>
        );
      }
      return (
        <ul className="space-y-3">
          {timeLogs.map((log) => (
            <TimeRow
              key={log.id}
              log={log}
              peopleNames={peopleNames}
              t={t}
              locale={i18n.language}
            />
          ))}
        </ul>
      );
    }
    const feed = buildFeed(comments, activity, timeLogs);
    return (
      <>
        {feed.length === 0 ? (
          <p className="text-flow-meta py-6 text-center" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.taskDetail.activityTabs.allEmpty')}
          </p>
        ) : (
          <ul className="mb-4 space-y-4">
            {feed.map((item) => {
              if (item.kind === 'comment') {
                return <li key={item.key}>{renderComments([item.comment])}</li>;
              }
              if (item.kind === 'history') {
                return (
                  <HistoryRow key={item.key} entry={item.entry} peopleNames={peopleNames} t={t} />
                );
              }
              return (
                <TimeRow
                  key={item.key}
                  log={item.log}
                  peopleNames={peopleNames}
                  t={t}
                  locale={i18n.language}
                />
              );
            })}
          </ul>
        )}
        {composer}
      </>
    );
  };

  return (
    <section aria-label={t('flow.taskDetail.activity')}>
      <div
        role="tablist"
        aria-label={t('flow.taskDetail.activity')}
        onKeyDown={onTabKeyDown}
        className="flex items-center gap-1 mb-4 overflow-x-auto"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        {ACTIVITY_TABS.map((name) => {
          const selected = tab === name;
          const count = counts[name];
          return (
            <button
              type="button"
              key={name}
              id={tabId(name)}
              role="tab"
              aria-selected={selected}
              aria-controls={panelId}
              tabIndex={selected ? 0 : -1}
              onClick={() => selectTab(name)}
              className="flow-focus shrink-0 px-2 py-2 text-flow-meta font-medium sm:px-3"
              style={{
                color: selected ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                borderBottom: selected ? '2px solid var(--accent)' : '2px solid transparent',
              }}
            >
              {t(TAB_LABEL_KEYS[name])}
              {count !== undefined && (
                <span
                  className="ms-1.5 text-flow-meta px-1 rounded-sm"
                  style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
                >
                  {count}
                </span>
              )}
            </button>
          );
        })}
      </div>
      <div id={panelId} role="tabpanel" aria-labelledby={tabId(tab)}>
        {renderPanel()}
      </div>
    </section>
  );
};

export default TaskActivityTabs;
