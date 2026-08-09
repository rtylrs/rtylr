import type { Task, Project } from '../stores/flowStore';

/** Returns archived tasks. */
export function archivedTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => t.archived === true);
}

/** Returns active (non-archived) tasks. */
export function activeTasks(tasks: Task[]): Task[] {
  return tasks.filter((t) => !t.archived);
}

/** Returns the count of archived tasks. */
export function archivedTaskCount(tasks: Task[]): number {
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
export function hasArchivedTasks(tasks: Task[]): boolean {
  return tasks.some((t) => t.archived === true);
}

/** Returns tasks archived from a specific project. */
export function archivedTasksForProject(tasks: Task[], projectId: string): Task[] {
  return tasks.filter((t) => t.archived === true && t.projectId === projectId);
}
