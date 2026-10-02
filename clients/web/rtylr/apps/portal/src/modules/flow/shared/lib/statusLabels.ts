import type { Task } from '../stores/flowStore';

/**
 * The i18n key for a task status.
 *
 * `locales/*​/flowStatus.json` shipped four of the six: `backlog` and `in_review`
 * had no entry, so anything rendering a status name for those two had nothing
 * to read. The two keys are added alongside this map.
 *
 * The irregular names are the snake_case ones, `in_progress` and `in_review`,
 * which the locale files spell in camelCase, so interpolating the raw status
 * does not resolve.
 */
const KEYS: Record<Task['status'], string> = {
  backlog: 'flowStatus.backlog',
  todo: 'flowStatus.todo',
  in_progress: 'flowStatus.inProgress',
  in_review: 'flowStatus.inReview',
  done: 'flowStatus.done',
  cancelled: 'flowStatus.cancelled',
};

export function statusLabelKey(status: Task['status']): string {
  return KEYS[status] ?? 'flowStatus.todo';
}
