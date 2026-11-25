import type { Task, TaskType } from '../stores/flowStore';
import type { ApiEpicRollup } from '../services/flowApi';

/**
 * Flow's hierarchy is epic > task > subtask (T-573). The server enforces it;
 * these helpers only decide what the UI offers, so a picker never lists a
 * parent the server would refuse.
 *
 * The type field is shared with T-380: epic is one value of it, next to task,
 * bug and feature.
 */

export const TASK_TYPES: readonly TaskType[] = ['task', 'feature', 'bug', 'chore', 'spike', 'epic'];

/** Stands in for "no epic" in the epic filter and the epic swimlane. */
export const NO_EPIC = 'none';

export function toTaskType(value: string | undefined | null): TaskType {
  return (TASK_TYPES as readonly string[]).includes(value ?? '') ? (value as TaskType) : 'task';
}

export function isEpic(task: Pick<Task, 'taskType'>): boolean {
  return task.taskType === 'epic';
}

/**
 * The epic a task belongs to for filtering and grouping. An epic belongs to
 * itself, so filtering by an epic keeps the epic's own card next to its tasks.
 */
export function epicKeyOf(task: Pick<Task, 'id' | 'taskType' | 'epicId'>): string {
  if (isEpic(task)) return task.id;
  return task.epicId ?? NO_EPIC;
}

export interface EpicSummary {
  id: string;
  title: string;
  color?: string;
}

/** The epics among a list of tasks, in list order. */
export function epicsIn(tasks: readonly Task[]): EpicSummary[] {
  return tasks
    .filter(isEpic)
    .map((task) => ({ id: task.id, title: task.title, color: task.epicColor }));
}

/**
 * Where a task may move to. Mirrors domain.PlaceUnderParent on the server:
 * the same project, never itself, an epic takes no parent, a parent must not
 * be a subtask, and a task that has subtasks may only go under an epic.
 */
export function parentCandidates(task: Task, tasks: readonly Task[]): Task[] {
  if (isEpic(task)) return [];
  return tasks.filter((candidate) => {
    if (candidate.id === task.id || candidate.archived) return false;
    if (candidate.projectId !== task.projectId) return false;
    if (candidate.parentId) return false;
    if (isEpic(candidate)) return candidate.id !== task.epicId || Boolean(task.parentId);
    if (task.subtaskCount > 0) return false;
    return candidate.id !== task.parentId;
  });
}

export type EpicProgressMode = 'count' | 'estimate' | 'hours';

export const EPIC_PROGRESS_MODES: readonly EpicProgressMode[] = ['count', 'estimate', 'hours'];

export interface EpicProgress {
  /** What has been done, in the mode's unit: tasks or minutes. */
  done: number;
  /** What it is measured against, in the same unit. Zero when nothing is. */
  total: number;
  /** 0 to 100, capped. Logged hours can pass the estimate; the bar stops full. */
  percent: number;
  /** True when logged hours have passed the estimate. */
  over: boolean;
}

/**
 * Progress of one epic.
 *
 * - count: done tasks out of the tasks still in play. Cancelled tasks are
 *   neither done nor left to do, so they leave the denominator.
 * - estimate: the estimate sitting on done tasks out of the whole estimate.
 * - hours: time logged out of the whole estimate.
 */
export function epicProgress(rollup: ApiEpicRollup, mode: EpicProgressMode): EpicProgress {
  const n = (v: number | undefined) => Number(v ?? 0);
  let done = 0;
  let total = 0;
  if (mode === 'count') {
    done = n(rollup.doneTasks);
    total = n(rollup.totalTasks) - n(rollup.cancelledTasks);
  } else if (mode === 'estimate') {
    done = n(rollup.doneEstimateMinutes);
    total = n(rollup.estimateMinutes);
  } else {
    done = n(rollup.loggedMinutes);
    total = n(rollup.estimateMinutes);
  }
  total = Math.max(0, total);
  const percent = total > 0 ? Math.min(100, Math.round((done / total) * 100)) : 0;
  return { done, total, percent, over: total > 0 && done > total };
}
