import type { Project } from '../stores/flowStore';

// Generic over any task-shaped object with an `archived` flag so these
// work on both the local-store `Task` and the wire `ApiTask` - the two
// task shapes flow through different pipelines but archive semantics
// (and the tests that pin them) should be identical for both.
interface Archivable {
  archived?: boolean;
}

/** Returns archived tasks. */
export function archivedTasks<T extends Archivable>(tasks: T[]): T[] {
  return tasks.filter((t) => t.archived === true);
}

/** Returns active (non-archived) tasks. */
export function activeTasks<T extends Archivable>(tasks: T[]): T[] {
  return tasks.filter((t) => !t.archived);
}

/** Returns the count of archived tasks. */
export function archivedTaskCount<T extends Archivable>(tasks: T[]): number {
  return archivedTasks(tasks).length;
}

/** Returns archived projects. */
export function archivedProjectList(projects: Project[]): Project[] {
  return projects.filter((p) => p.archived === true);
}

/** Returns active (non-archived) projects. */
export function activeProjectList(projects: Project[]): Project[] {
  return projects.filter((p) => !p.archived);
}

/** Returns true if any task is archived. */
export function hasArchivedTasks<T extends Archivable>(tasks: T[]): boolean {
  return tasks.some((t) => t.archived === true);
}

/** Returns tasks archived from a specific project. */
export function archivedTasksForProject<T extends Archivable & { projectId: string }>(
  tasks: T[],
  projectId: string
): T[] {
  return tasks.filter((t) => t.archived === true && t.projectId === projectId);
}
