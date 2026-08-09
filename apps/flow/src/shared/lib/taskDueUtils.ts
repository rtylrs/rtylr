import type { Task } from '../stores/flowStore';

/** Returns tasks due today. */
export function tasksDueToday(tasks: Task[], today: string): Task[] {
  return tasks.filter((t) => t.dueDate === today);
}

/** Returns tasks due within N days from today (inclusive). */
export function tasksDueWithin(tasks: Task[], today: string, days: number): Task[] {
  const cutoff = addDays(today, days);
  return tasks.filter(
    (t) =>
      t.dueDate &&
      t.dueDate >= today &&
      t.dueDate <= cutoff &&
      t.status !== 'done' &&
      t.status !== 'cancelled'
  );
}

/** Returns tasks overdue (past due, not done or cancelled). */
export function overdueTaskList(tasks: Task[], today: string): Task[] {
  return tasks.filter(
    (t) => t.dueDate && t.dueDate < today && t.status !== 'done' && t.status !== 'cancelled'
  );
}

/** Returns tasks with no due date (and not done/cancelled). */
export function tasksNoDueDate(tasks: Task[]): Task[] {
  return tasks.filter((t) => !t.dueDate && t.status !== 'done' && t.status !== 'cancelled');
}

/** Returns tasks due this week (Monday to Sunday). */
export function tasksDueThisWeek(tasks: Task[], today: string): Task[] {
  const monday = startOfWeek(today);
  const sunday = addDays(monday, 6);
  return tasks.filter((t) => t.dueDate && t.dueDate >= monday && t.dueDate <= sunday);
}

/** Returns the due status label for a task. */
export function dueDateLabel(task: Task, today: string): string {
  if (!task.dueDate) return 'No due date';
  if (task.dueDate < today) return 'Overdue';
  if (task.dueDate === today) return 'Due today';
  const diff = daysBetween(today, task.dueDate);
  if (diff === 1) return 'Due tomorrow';
  if (diff <= 7) return `Due in ${diff} days`;
  return `Due ${task.dueDate}`;
}

function daysBetween(a: string, b: string): number {
  return Math.round((new Date(b).getTime() - new Date(a).getTime()) / 86_400_000);
}

function addDays(date: string, days: number): string {
  const [y, m, d] = date.split('-').map(Number);
  const result = new Date(Date.UTC(y, m - 1, d + days));
  return result.toISOString().slice(0, 10);
}

function startOfWeek(date: string): string {
  const [y, m, d] = date.split('-').map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d));
  const day = dt.getUTCDay();
  const diff = day === 0 ? -6 : 1 - day;
  dt.setUTCDate(dt.getUTCDate() + diff);
  return dt.toISOString().slice(0, 10);
}
