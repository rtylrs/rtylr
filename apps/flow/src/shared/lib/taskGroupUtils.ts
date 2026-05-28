import type { Task } from '../stores/flowStore';

/** Group tasks by project ID. */
export function groupByProject(tasks: Task[]): Map<string, Task[]> {
  const map = new Map<string, Task[]>();
  for (const t of tasks) {
    const bucket = map.get(t.projectId) ?? [];
    bucket.push(t);
    map.set(t.projectId, bucket);
  }
  return map;
}

/** Group tasks by assignee ID ('unassigned' for tasks with no assignee). */
export function groupByAssignee(tasks: Task[]): Map<string, Task[]> {
  const map = new Map<string, Task[]>();
  for (const t of tasks) {
    const key = t.assigneeId || 'unassigned';
    const bucket = map.get(key) ?? [];
    bucket.push(t);
    map.set(key, bucket);
  }
  return map;
}

/** Group tasks by cycle ID ('none' for tasks without a cycle). */
export function groupByCycle(tasks: Task[]): Map<string, Task[]> {
  const map = new Map<string, Task[]>();
  for (const t of tasks) {
    const key = t.cycleId ?? 'none';
    const bucket = map.get(key) ?? [];
    bucket.push(t);
    map.set(key, bucket);
  }
  return map;
}

/** Group tasks by milestone ID ('none' for tasks without a milestone). */
export function groupByMilestone(tasks: Task[]): Map<string, Task[]> {
  const map = new Map<string, Task[]>();
  for (const t of tasks) {
    const key = t.milestoneId ?? 'none';
    const bucket = map.get(key) ?? [];
    bucket.push(t);
    map.set(key, bucket);
  }
  return map;
}

/** Group tasks by parent task ID ('root' for top-level tasks). */
export function groupByParent(tasks: Task[]): Map<string, Task[]> {
  const map = new Map<string, Task[]>();
  for (const t of tasks) {
    const key = t.parentId ?? 'root';
    const bucket = map.get(key) ?? [];
    bucket.push(t);
    map.set(key, bucket);
  }
  return map;
}

/** Returns root-level tasks (tasks without a parent). */
export function rootTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => !t.parentId);
}

/** Returns child tasks of a given parent task. */
export function childTasks(tasks: Task[], parentId: string): Task[] {
  return tasks.filter((t) => t.parentId === parentId);
}
