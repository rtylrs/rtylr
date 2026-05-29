import type { Task } from '../stores/flowStore';

export interface TaskFilter {
  status?: Task['status'][];
  priority?: Task['priority'][];
  assigneeId?: string;
  projectId?: string;
  labelIds?: string[];
  cycleId?: string;
  milestoneId?: string;
  archived?: boolean;
  hasDueDate?: boolean;
  hasEstimate?: boolean;
}

/** Returns true if the task passes all filter criteria. */
export function taskPassesFilter(task: Task, filter: TaskFilter): boolean {
  if (filter.status && filter.status.length > 0 && !filter.status.includes(task.status)) {
    return false;
  }
  if (filter.priority && filter.priority.length > 0 && !filter.priority.includes(task.priority)) {
    return false;
  }
  if (filter.assigneeId !== undefined && task.assigneeId !== filter.assigneeId) {
    return false;
  }
  if (filter.projectId !== undefined && task.projectId !== filter.projectId) {
    return false;
  }
  if (filter.labelIds && filter.labelIds.length > 0) {
    const taskLabels = task.labelIds ?? [];
    const hasLabel = filter.labelIds.some((id) => taskLabels.includes(id));
    if (!hasLabel) return false;
  }
  if (filter.cycleId !== undefined && task.cycleId !== filter.cycleId) {
    return false;
  }
  if (filter.milestoneId !== undefined && task.milestoneId !== filter.milestoneId) {
    return false;
  }
  if (filter.archived !== undefined && (task.archived ?? false) !== filter.archived) {
    return false;
  }
  if (filter.hasDueDate === true && !task.dueDate) {
    return false;
  }
  if (filter.hasDueDate === false && !!task.dueDate) {
    return false;
  }
  if (filter.hasEstimate === true && task.estimate == null) {
    return false;
  }
  if (filter.hasEstimate === false && task.estimate != null) {
    return false;
  }
  return true;
}

/** Applies a TaskFilter to a list of tasks and returns matching tasks. */
export function applyFilter(tasks: Task[], filter: TaskFilter): Task[] {
  return tasks.filter((t) => taskPassesFilter(t, filter));
}

/** Merges two filters (second overrides first). */
export function mergeFilters(base: TaskFilter, override: TaskFilter): TaskFilter {
  return { ...base, ...override };
}

/** Returns true if the filter has any active criteria. */
export function hasActiveFilter(filter: TaskFilter): boolean {
  return activeFilterCount(filter) > 0;
}

/** Returns the number of active filter criteria. */
export function activeFilterCount(filter: TaskFilter): number {
  let count = 0;
  if ((filter.status?.length ?? 0) > 0) count++;
  if ((filter.priority?.length ?? 0) > 0) count++;
  if (filter.assigneeId !== undefined) count++;
  if (filter.projectId !== undefined) count++;
  if ((filter.labelIds?.length ?? 0) > 0) count++;
  if (filter.cycleId !== undefined) count++;
  if (filter.milestoneId !== undefined) count++;
  if (filter.archived !== undefined) count++;
  if (filter.hasDueDate !== undefined) count++;
  if (filter.hasEstimate !== undefined) count++;
  return count;
}

/** Returns an empty filter object. */
export function emptyFilter(): TaskFilter {
  return {};
}
