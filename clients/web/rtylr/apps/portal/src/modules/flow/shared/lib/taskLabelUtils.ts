import type { Label, Task } from '../stores/flowStore';

/**
 * Returns labels assigned to a task.
 *
 * Ids are compared as strings. `Label.id` from the store is a string, but the
 * same shape arrives from `useLabels` as `ApiLabel`, which declares `id: number`
 * while the wire actually sends the int64 as a string, and `Array.includes`
 * does not coerce. The label parameter is widened for the same reason, so a
 * caller holding API labels does not need a cast.
 */
export function labelsForTask<T extends { id: string | number }>(labels: T[], task: Task): T[] {
  const ids = (task.labelIds ?? []).map(String);
  return labels.filter((l) => ids.includes(String(l.id)));
}

/** Returns tasks that have a specific label. */
export function tasksWithLabel(tasks: Task[], labelId: string): Task[] {
  return tasks.filter((t) => (t.labelIds ?? []).includes(labelId));
}

/** Returns tasks that have any of the given label IDs. */
export function tasksWithAnyLabel(tasks: Task[], labelIds: string[]): Task[] {
  if (labelIds.length === 0) return tasks;
  const set = new Set(labelIds);
  return tasks.filter((t) => (t.labelIds ?? []).some((id) => set.has(id)));
}

/** Returns tasks that have all of the given label IDs. */
export function tasksWithAllLabels(tasks: Task[], labelIds: string[]): Task[] {
  if (labelIds.length === 0) return tasks;
  return tasks.filter((t) => {
    const taskLabels = new Set(t.labelIds ?? []);
    return labelIds.every((id) => taskLabels.has(id));
  });
}

/** Returns tasks that have no labels. */
export function unlabeledTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => !t.labelIds || t.labelIds.length === 0);
}

/** Returns tasks that have at least one label. */
export function labeledTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => (t.labelIds ?? []).length > 0);
}

/** Returns the count of tasks per label. */
export function taskCountByLabel(tasks: Task[], labels: Label[]): Map<string, number> {
  const map = new Map<string, number>(labels.map((l) => [l.id, 0]));
  for (const task of tasks) {
    for (const id of task.labelIds ?? []) {
      if (map.has(id)) map.set(id, (map.get(id) ?? 0) + 1);
    }
  }
  return map;
}

/** Returns labels sorted by number of uses descending. */
export function labelsByUsage(labels: Label[], tasks: Task[]): Label[] {
  const counts = taskCountByLabel(tasks, labels);
  return [...labels].sort((a, b) => (counts.get(b.id) ?? 0) - (counts.get(a.id) ?? 0));
}
