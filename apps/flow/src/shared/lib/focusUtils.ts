import type { Task } from '../stores/flowStore';

/** Returns tasks that are pinned to the focus/today view. */
export function focusPinnedTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => t.focusPin);
}

/** Returns tasks due today (ISO date string format). */
export function tasksDueToday(tasks: Task[], today: string): Task[] {
  return tasks.filter((t) => t.dueDate === today);
}

/** Returns tasks due today or pinned to focus. */
export function focusTodayTasks(tasks: Task[], today: string): Task[] {
  const seen = new Set<string>();
  const result: Task[] = [];
  for (const t of tasks) {
    if ((t.dueDate === today || t.focusPin) && !seen.has(t.id)) {
      seen.add(t.id);
      result.push(t);
    }
  }
  return result;
}

/** Returns tasks that are in a waiting state. */
export function waitingTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => !!t.waitingOn);
}

/** Returns tasks that are in progress and not waiting. */
export function activeFocusTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => t.status === 'in_progress' && !t.waitingOn);
}

/** Returns tasks within effortMinutes budget (for small win picking). */
export function tasksWithinEffortBudget(tasks: Task[], budgetMinutes: number): Task[] {
  return tasks.filter(
    (t) =>
      t.effortMinutes !== undefined &&
      t.effortMinutes > 0 &&
      t.effortMinutes <= budgetMinutes &&
      t.status !== 'done' &&
      t.status !== 'cancelled'
  );
}

/** Returns tasks that have a decision note. */
export function tasksWithDecisionNote(tasks: Task[]): Task[] {
  return tasks.filter((t) => !!t.decisionNote);
}
