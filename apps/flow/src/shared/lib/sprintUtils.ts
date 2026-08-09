import type { Cycle, Task } from '../stores/flowStore';

export interface SprintVelocity {
  cycleId: string;
  cycleName: string;
  completed: number;
  scope: number;
}

/** Computes velocity data for each cycle. */
export function sprintVelocities(cycles: Cycle[]): SprintVelocity[] {
  return cycles.map((c) => ({
    cycleId: c.id,
    cycleName: c.name,
    completed: c.completed,
    scope: c.scope,
  }));
}

/** Returns average completed tasks per cycle (simple velocity). */
export function averageVelocity(cycles: Cycle[]): number {
  const completed = cycles.filter((c) => c.status === 'completed');
  if (completed.length === 0) return 0;
  const total = completed.reduce((sum, c) => sum + c.completed, 0);
  return Math.round((total / completed.length) * 10) / 10;
}

/** Returns the scope variance for a cycle: completed - scope (positive = over-delivered). */
export function scopeVariance(cycle: Cycle): number {
  return cycle.completed - cycle.scope;
}

/** Returns the predictability ratio: completed / max(scope, completed). */
export function predictability(cycle: Cycle): number {
  const max = Math.max(cycle.scope, cycle.completed);
  if (max === 0) return 1;
  return Math.round((cycle.completed / max) * 100) / 100;
}

/** Returns incomplete tasks carried over from a cycle. */
export function carryoverTasks(tasks: Task[], cycleId: string): Task[] {
  return tasks.filter(
    (t) => t.cycleId === cycleId && t.status !== 'done' && t.status !== 'cancelled'
  );
}

/** Returns the carryover rate (0-1) for a cycle based on its task list. */
export function carryoverRate(tasks: Task[], cycle: Cycle): number {
  if (cycle.scope === 0) return 0;
  const carried = carryoverTasks(tasks, cycle.id).length;
  return Math.round((carried / cycle.scope) * 100) / 100;
}

/** Returns cycles sorted by start date descending (most recent first). */
export function sortCyclesDesc(cycles: Cycle[]): Cycle[] {
  return [...cycles].sort((a, b) => b.startDate.localeCompare(a.startDate));
}
