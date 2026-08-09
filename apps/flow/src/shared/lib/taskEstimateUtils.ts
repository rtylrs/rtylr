import type { Task } from '../stores/flowStore';

/** Returns true if a task has an estimate set. */
export function hasEstimate(task: Task): boolean {
  return task.estimate != null && task.estimate > 0;
}

/** Returns tasks that have an estimate. */
export function estimatedTasks(tasks: Task[]): Task[] {
  return tasks.filter(hasEstimate);
}

/** Returns tasks that do not have an estimate. */
export function unestimatedTaskList(tasks: Task[]): Task[] {
  return tasks.filter((t) => !hasEstimate(t));
}

/** Returns the estimated hours for a task, or 0 if not set. */
export function estimateHours(task: Task): number {
  return task.estimate ?? 0;
}

/** Returns the total estimated hours for a list of tasks. */
export function totalEstimate(tasks: Task[]): number {
  return tasks.reduce((sum, t) => sum + estimateHours(t), 0);
}

/** Returns the average estimate across tasks that have one. */
export function averageEstimate(tasks: Task[]): number {
  const withEstimate = estimatedTasks(tasks);
  if (withEstimate.length === 0) return 0;
  return Math.round((totalEstimate(withEstimate) / withEstimate.length) * 10) / 10;
}

/** Returns tasks sorted by estimate descending (largest first). */
export function sortByEstimateDesc(tasks: Task[]): Task[] {
  return [...tasks].sort((a, b) => estimateHours(b) - estimateHours(a));
}

/** Returns tasks within the given estimate range (inclusive). */
export function tasksWithEstimateRange(tasks: Task[], minHours: number, maxHours: number): Task[] {
  return tasks.filter((t) => {
    const h = estimateHours(t);
    return h >= minHours && h <= maxHours;
  });
}

/** Returns the percentage of tasks with estimates. */
export function estimateCoverage(tasks: Task[]): number {
  if (tasks.length === 0) return 0;
  return Math.round((estimatedTasks(tasks).length / tasks.length) * 100);
}
