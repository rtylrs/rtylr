import type { Task, Project, Member } from '../stores/flowStore';

/** Returns the count of tasks assigned to a specific member. */
export function taskCountForMember(tasks: Task[], memberId: string): number {
  return tasks.filter((t) => t.assigneeId === memberId).length;
}

/** Returns the count of tasks in a specific project. */
export function taskCountForProject(tasks: Task[], projectId: string): number {
  return tasks.filter((t) => t.projectId === projectId).length;
}

/** Returns the number of tasks in each project as a map. */
export function taskCountsPerProject(tasks: Task[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const task of tasks) {
    map.set(task.projectId, (map.get(task.projectId) ?? 0) + 1);
  }
  return map;
}

/** Returns the number of tasks per assignee as a map. */
export function taskCountsPerMember(tasks: Task[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const task of tasks) {
    if (!task.assigneeId) continue;
    map.set(task.assigneeId, (map.get(task.assigneeId) ?? 0) + 1);
  }
  return map;
}

/** Returns the project with the most tasks. */
export function projectWithMostTasks(projects: Project[]): Project | null {
  if (projects.length === 0) return null;
  return projects.reduce((max, p) => (p.taskCount > max.taskCount ? p : max), projects[0]);
}

/** Returns the member with the most assigned tasks. */
export function memberWithMostTasks(tasks: Task[], members: Member[]): Member | null {
  if (members.length === 0) return null;
  const counts = taskCountsPerMember(tasks);
  return members.reduce(
    (max, m) => ((counts.get(m.id) ?? 0) > (counts.get(max.id) ?? 0) ? m : max),
    members[0]
  );
}

/** Returns the total number of open tasks (not done or cancelled). */
export function openTaskCount(tasks: Task[]): number {
  return tasks.filter((t) => t.status !== 'done' && t.status !== 'cancelled').length;
}

/** Returns the ratio of completed to total tasks (0-1). */
export function taskCompletionRatio(tasks: Task[]): number {
  if (tasks.length === 0) return 0;
  const done = tasks.filter((t) => t.status === 'done' || t.status === 'cancelled').length;
  return done / tasks.length;
}
