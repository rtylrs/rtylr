import type { Task } from '../stores/flowStore';

/** Returns tasks in backlog status. */
export function backlogTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => t.status === 'backlog');
}

/** Returns tasks with no assignee. */
export function unassignedTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => !t.assigneeId);
}

/** Returns tasks with no due date set. */
export function tasksMissingDueDate(tasks: Task[]): Task[] {
  return tasks.filter((t) => !t.dueDate);
}

/** Returns tasks with no_priority priority. */
export function unprioritizedTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => t.priority === 'no_priority');
}

/** Returns tasks with no estimate set. */
export function unestimatedTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => t.estimate == null);
}

/**
 * Returns tasks that need triage: backlog OR missing one of
 * (assignee, due date, priority). Done/cancelled tasks are excluded.
 */
export function triageNeeded(tasks: Task[]): Task[] {
  return tasks.filter((t) => {
    if (t.status === 'done' || t.status === 'cancelled') return false;
    return t.status === 'backlog' || !t.assigneeId || !t.dueDate || t.priority === 'no_priority';
  });
}

/**
 * Returns a score 0-4 indicating how much triage a task needs
 * (one point each for: backlog status, no assignee, no due date, no priority).
 */
export function triageScore(task: Task): number {
  let score = 0;
  if (task.status === 'backlog') score++;
  if (!task.assigneeId) score++;
  if (!task.dueDate) score++;
  if (task.priority === 'no_priority') score++;
  return score;
}

/** Sorts tasks by triage score descending (most incomplete first). */
export function sortByTriageScore(tasks: Task[]): Task[] {
  return [...tasks].sort((a, b) => triageScore(b) - triageScore(a));
}
