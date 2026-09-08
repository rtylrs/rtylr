import type { Task } from '../stores/flowStore';
import { todayIso, addDays } from './dateUtils';

/**
 * Urgency buckets for "My tasks".
 *
 * Five sections in urgency order, which S-04 and Panel B both specify. The
 * previous four folded **overdue into Today**, which is the defect rtylr#1077
 * was raised for: with six overdue tasks assigned, nothing promoted them, so a
 * fifteen-day-overdue item sat in the middle of a section headed "Today".
 *
 * `no_date` replaces the old `recently_assigned`. That bucket keyed off
 * `createdAt` and answered a different question, "what landed on me lately",
 * which is not a position on an urgency scale: a task assigned this morning and
 * due in a month sorted above one due tomorrow.
 */
export type MyTaskBucket = 'late' | 'today' | 'this_week' | 'later' | 'no_date';

/**
 * Compared as strings, never as Dates.
 *
 * A due date is a calendar day, not an instant. Parsing `"2026-09-08"` with
 * `new Date()` gives UTC midnight, and reading that back locally names the day
 * before for anyone west of Greenwich, which is what put the project calendar a
 * day out (rtylr#1215). The previous version of this file did exactly that:
 * `startOfDay(new Date(task.dueDate))`.
 *
 * `todayIso` and `addDays` are both local-in, local-out, and `YYYY-MM-DD`
 * strings sort lexicographically in date order, so `<` and `>` are the whole
 * comparison.
 */
export function bucketForMyTask(task: Task): MyTaskBucket {
  const due = task.dueDate ? task.dueDate.split('T')[0] : '';
  if (due === '') return 'no_date';

  const today = todayIso();
  if (due < today) return 'late';
  if (due === today) return 'today';
  return due <= addDays(today, 7) ? 'this_week' : 'later';
}

/**
 * Translation keys, not English. These were literal strings while nothing
 * imported the module; rendering them as-is would have put English section
 * headings into a screen that ships in three languages.
 */
export const MY_TASK_BUCKET_LABEL_KEY: Record<MyTaskBucket, string> = {
  late: 'flow.myIssues.buckets.late',
  today: 'flow.myIssues.buckets.today',
  this_week: 'flow.myIssues.buckets.thisWeek',
  later: 'flow.myIssues.buckets.later',
  no_date: 'flow.myIssues.buckets.noDate',
};

/** Urgency order, which is the reading order of the screen. */
export const MY_TASK_BUCKET_ORDER: MyTaskBucket[] = [
  'late',
  'today',
  'this_week',
  'later',
  'no_date',
];

/**
 * Which buckets open collapsed.
 *
 * Late and Today never do: the point of the screen is that overdue work is
 * visible without an interaction. `no_date` does, because it is the bucket with
 * no claim on the reader's attention and, on a busy account, the longest.
 */
export const MY_TASK_BUCKET_COLLAPSED_BY_DEFAULT: ReadonlySet<MyTaskBucket> = new Set<MyTaskBucket>(
  ['no_date']
);
