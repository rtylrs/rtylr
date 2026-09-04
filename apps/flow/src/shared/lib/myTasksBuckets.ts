import type { Task } from '../stores/flowStore';

export type MyTaskBucket = 'recently_assigned' | 'today' | 'upcoming' | 'later';

function startOfDay(d: Date): Date {
  const x = new Date(d);
  x.setHours(0, 0, 0, 0);
  return x;
}

function addDays(d: Date, n: number): Date {
  const x = new Date(d);
  x.setDate(x.getDate() + n);
  return x;
}

/**
 * Time-based buckets for "My tasks": Recently assigned, Today, Upcoming, Later.
 * Overdue tasks are grouped under Today.
 */
export function bucketForMyTask(task: Task): MyTaskBucket {
  const now = new Date();
  const today = startOfDay(now);
  const weekEnd = addDays(today, 7);

  if (!task.dueDate || task.dueDate === '') {
    const created = new Date(task.createdAt);
    const weekAgo = addDays(today, -7);
    if (created >= weekAgo) return 'recently_assigned';
    return 'later';
  }

  const due = startOfDay(new Date(task.dueDate));
  if (due < today) return 'today';
  if (due.getTime() === today.getTime()) return 'today';
  if (due > today && due <= weekEnd) return 'upcoming';
  return 'later';
}

/**
 * Translation keys, not English. These were literal strings while nothing
 * imported the module; rendering them as-is would have put English section
 * headings into a screen that ships in three languages.
 */
export const MY_TASK_BUCKET_LABEL_KEY: Record<MyTaskBucket, string> = {
  recently_assigned: 'flow.myIssues.buckets.recentlyAssigned',
  today: 'flow.myIssues.buckets.today',
  upcoming: 'flow.myIssues.buckets.upcoming',
  later: 'flow.myIssues.buckets.later',
};

export const MY_TASK_BUCKET_ORDER: MyTaskBucket[] = [
  'recently_assigned',
  'today',
  'upcoming',
  'later',
];
