import type { Task } from '../stores/flowStore';

/** All valid task statuses in display order. */
export const TASK_STATUSES: Task['status'][] = [
  'backlog',
  'todo',
  'in_progress',
  'in_review',
  'done',
  'cancelled',
] as const;

/** All valid task priorities in urgency order. */
export const TASK_PRIORITIES: Task['priority'][] = [
  'urgent',
  'high',
  'medium',
  'low',
  'no_priority',
] as const;

/** Statuses that count as "open" work. */
export const OPEN_STATUSES: Task['status'][] = ['backlog', 'todo', 'in_progress', 'in_review'];

/** Statuses that count as "closed" work. */
export const CLOSED_STATUSES: Task['status'][] = ['done', 'cancelled'];

/** Statuses that represent active work in flight. */
export const ACTIVE_STATUSES: Task['status'][] = ['in_progress', 'in_review'];

/**
 * i18n key for each task status's display label (all under the `flow`
 * locale namespace's `taskDetail.*` path - see locales/en/flow.json).
 * This is the single source of truth for status text: every view that
 * shows a status label must resolve through this map instead of hardcoding
 * its own string, so two views can never drift apart again (issue #756,
 * where the project List view hardcoded "Todo" while the rest of the app
 * used the i18n-backed "To Do").
 */
export const STATUS_LABEL_KEYS: Record<Task['status'], string> = {
  backlog: 'flow.taskDetail.backlog',
  todo: 'flow.taskDetail.todo',
  in_progress: 'flow.taskDetail.inProgress',
  in_review: 'flow.taskDetail.inReview',
  done: 'flow.taskDetail.done',
  cancelled: 'flow.taskDetail.cancelled',
};

/** Default page size for task lists. */
export const DEFAULT_PAGE_SIZE = 25;

/** Maximum allowed task title length. */
export const MAX_TASK_TITLE_LENGTH = 255;

/** Maximum allowed project name length. */
export const MAX_PROJECT_NAME_LENGTH = 100;

/** Effort bucket sizes in minutes for quick selection. */
export const EFFORT_BUCKETS_MINUTES = [15, 30, 60, 120, 240] as const;

/** Default cycle length in days. */
export const DEFAULT_CYCLE_DAYS = 14;

/** Health statuses for projects. */
export const PROJECT_HEALTH_OPTIONS = ['on_track', 'at_risk', 'off_track'] as const;

/** Milestone statuses. */
export const MILESTONE_STATUSES = ['planning', 'active', 'at_risk', 'completed'] as const;
