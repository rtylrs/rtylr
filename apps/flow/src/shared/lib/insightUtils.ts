import type { Task, Project, Member, TimeLog } from '../stores/flowStore';

/** Returns tasks assigned to a specific member. */
export function tasksForMember(tasks: Task[], memberId: string): Task[] {
  return tasks.filter((t) => t.assigneeId === memberId);
}

/** Counts done tasks for a member. */
export function completedTasksForMember(tasks: Task[], memberId: string): number {
  return tasksForMember(tasks, memberId).filter((t) => t.status === 'done').length;
}

/** Returns completion rate (0-100) for a member. */
export function memberCompletionRate(tasks: Task[], memberId: string): number {
  const assigned = tasksForMember(tasks, memberId);
  if (assigned.length === 0) return 0;
  const done = assigned.filter((t) => t.status === 'done').length;
  return Math.round((done / assigned.length) * 100);
}

/** Returns total logged minutes for a member across all timelogs. */
export function totalLoggedMinutesForMember(timelogs: TimeLog[], memberId: string): number {
  return timelogs.filter((l) => l.userId === memberId).reduce((sum, l) => sum + l.minutes, 0);
}

/** Returns average task estimate (hours) for a project's tasks. */
export function averageEstimateHours(tasks: Task[], projectId: string): number {
  const projectTasks = tasks.filter((t) => t.projectId === projectId && t.estimate != null);
  if (projectTasks.length === 0) return 0;
  const total = projectTasks.reduce((sum, t) => sum + (t.estimate ?? 0), 0);
  return Math.round((total / projectTasks.length) * 10) / 10;
}

/** Returns the number of tasks per member as a map (memberId -> count). */
export function taskCountByMember(tasks: Task[]): Map<string, number> {
  const map = new Map<string, number>();
  for (const task of tasks) {
    if (!task.assigneeId) continue;
    map.set(task.assigneeId, (map.get(task.assigneeId) ?? 0) + 1);
  }
  return map;
}

/** Returns team members sorted by assigned task count descending. */
export function membersByWorkload(tasks: Task[], members: Member[]): Member[] {
  const counts = taskCountByMember(tasks);
  return [...members].sort((a, b) => (counts.get(b.id) ?? 0) - (counts.get(a.id) ?? 0));
}

/** Returns projects sorted by task completion percentage descending. */
export function projectsByCompletion(projects: Project[]): Project[] {
  return [...projects].sort((a, b) => {
    const pctA = a.taskCount === 0 ? 0 : a.completedCount / a.taskCount;
    const pctB = b.taskCount === 0 ? 0 : b.completedCount / b.taskCount;
    return pctB - pctA;
  });
}

/** Returns the number of overdue tasks (past due date, not done or cancelled). */
export function overdueCount(tasks: Task[], today: string): number {
  return tasks.filter(
    (t) => t.dueDate && t.dueDate < today && t.status !== 'done' && t.status !== 'cancelled'
  ).length;
}

/** Returns the percentage of tasks with an estimate set. */
export function estimatedTaskPercent(tasks: Task[]): number {
  if (tasks.length === 0) return 0;
  const withEstimate = tasks.filter((t) => t.estimate != null).length;
  return Math.round((withEstimate / tasks.length) * 100);
}
