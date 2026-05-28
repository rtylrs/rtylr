import type { Task } from '../stores/flowStore';

/** Returns tasks that the given task depends on. */
export function dependenciesOf(tasks: Task[], taskId: string): Task[] {
  const task = tasks.find((t) => t.id === taskId);
  if (!task || !task.dependencies || task.dependencies.length === 0) return [];
  const depIds = new Set(task.dependencies);
  return tasks.filter((t) => depIds.has(t.id));
}

/** Returns tasks that depend on (block on) the given task. */
export function dependentsOf(tasks: Task[], taskId: string): Task[] {
  return tasks.filter((t) => t.dependencies?.includes(taskId));
}

/** Returns true when a task has any dependencies. */
export function hasDependencies(task: Task): boolean {
  return !!(task.dependencies && task.dependencies.length > 0);
}

/** Returns true when a task is blocking other tasks. */
export function isBlocking(tasks: Task[], taskId: string): boolean {
  return tasks.some((t) => t.dependencies?.includes(taskId));
}

/** Returns true when all dependencies of a task are done or cancelled. */
export function allDependenciesSatisfied(tasks: Task[], taskId: string): boolean {
  const deps = dependenciesOf(tasks, taskId);
  if (deps.length === 0) return true;
  return deps.every((d) => d.status === 'done' || d.status === 'cancelled');
}

/** Returns tasks whose dependencies are all satisfied (ready to start). */
export function readyTasks(tasks: Task[]): Task[] {
  return tasks.filter(
    (t) => t.status !== 'done' && t.status !== 'cancelled' && allDependenciesSatisfied(tasks, t.id)
  );
}

/** Returns tasks that are blocked (have unmet dependencies). */
export function blockedTasks(tasks: Task[]): Task[] {
  return tasks.filter(
    (t) =>
      t.status !== 'done' &&
      t.status !== 'cancelled' &&
      hasDependencies(t) &&
      !allDependenciesSatisfied(tasks, t.id)
  );
}
