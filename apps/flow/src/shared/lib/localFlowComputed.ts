import type { DashboardStats, ApiInsights, ApiTeamWorkload, ApiTask } from '../services/flowApi';
import type { Task, Project, Member } from '../stores/flowStore';
import { isTaskMine } from './currentUser';

export function buildDashboardStats(
  workspaceId: string,
  tasks: Task[],
  projects: Project[]
): DashboardStats {
  const wsTasks = tasks.filter((t) => t.workspaceId === workspaceId && !t.archived);
  const now = new Date();
  const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const myTasks = wsTasks
    .filter((t) => isTaskMine(t.assigneeId) && t.status !== 'done' && t.status !== 'cancelled')
    .slice(0, 12);

  const openTasks = wsTasks.filter((t) => t.status === 'todo' || t.status === 'backlog').length;
  const inProgress = wsTasks.filter((t) => t.status === 'in_progress' || t.status === 'in_review').length;
  const completedThisWeek = wsTasks.filter((t) => {
    if (t.status !== 'done') return false;
    return new Date(t.updatedAt) >= weekAgo;
  }).length;
  const overdue = wsTasks.filter((t) => {
    if (!t.dueDate || t.status === 'done') return false;
    return new Date(t.dueDate) < now;
  }).length;

  const recentProjects = projects
    .filter((p) => p.workspaceId === workspaceId)
    .slice(0, 6);

  return {
    openTasks,
    inProgress,
    completedThisWeek,
    overdue,
    myTasks: myTasks as unknown as ApiTask[],
    recentProjects: recentProjects as any,
  };
}

export function buildInsights(workspaceId: string, tasks: Task[]): ApiInsights {
  const ws = tasks.filter((t) => t.workspaceId === workspaceId);
  const totalTasks = ws.length;
  const doneTasks = ws.filter((t) => t.status === 'done').length;

  const velocity = [
    { cycle: 'C-1', completed: 4, total: 12, completionRate: 0.33 },
    { cycle: 'C-2', completed: 7, total: 14, completionRate: 0.5 },
    { cycle: 'C-3', completed: 6, total: 11, completionRate: 0.55 },
  ];

  const burndown = [0, 1, 2, 3, 4, 5, 6, 7].map((day) => ({
    day,
    ideal: 14 - day * 2,
    actual: Math.max(0, 14 - day * 2 + (day % 3) - 1),
  }));

  const timeByMember = [
    { id: 'u1', name: 'Alex', hours: 32 },
    { id: 'u2', name: 'Sam', hours: 24 },
  ];

  const throughput = ['W1', 'W2', 'W3', 'W4', 'W5', 'W6'].map((week, i) => ({
    week,
    completed: 3 + (i % 4),
  }));

  return {
    velocity,
    burndown,
    cycleTime: 4,
    timeByMember,
    throughput,
    totalTasks,
    doneTasks,
  };
}

export function buildTeamWorkload(members: Member[], tasks: Task[]): ApiTeamWorkload[] {
  return members.map((m) => {
    const mine = tasks.filter((t) => t.assigneeId === m.id && !t.archived);
    const assignedHours = mine.reduce((s, t) => s + (t.estimate ?? 2), 0);
    const load = m.capacity > 0 ? Math.min(100, Math.round((assignedHours / m.capacity) * 100)) : 0;
    return {
      id: m.id,
      name: m.name,
      initials: m.initials,
      capacity: m.capacity,
      assignedHours,
      load,
      tasks: mine as unknown as ApiTask[],
    };
  });
}
