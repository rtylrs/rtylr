import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useProjectOverview(projectId: string | undefined) {
  const { projects, tasks, milestones, members, comments } = useFlowStore();
  return useMemo(() => {
    if (!projectId) return null;
    const project = projects.find((p) => p.id === projectId);
    if (!project) return null;
    const pTasks = tasks.filter((t) => t.projectId === projectId);
    const done = pTasks.filter((t) => t.status === 'done').length;
    const inProgress = pTasks.filter((t) => t.status === 'in_progress').length;
    const blocked = pTasks.filter(
      (t) => t.dependencies && t.dependencies.length > 0
    ).length;
    const overdue = pTasks.filter((t) => {
      if (!t.dueDate) return false;
      return new Date(t.dueDate) < new Date() && t.status !== 'done';
    }).length;
    const pMilestones = milestones.filter((m) => m.projectId === projectId);
    const teamMembers = members.filter((m) =>
      project.memberAvatars.includes(m.initials)
    );
    const recentComments = comments
      .filter((c) => pTasks.some((t) => t.id === c.taskId))
      .sort((a, b) => b.createdAt.localeCompare(a.createdAt))
      .slice(0, 5);
    return {
      project,
      tasks: pTasks,
      stats: {
        total: pTasks.length,
        done,
        inProgress,
        blocked,
        overdue,
        progress: pTasks.length ? Math.round((done / pTasks.length) * 100) : 0,
      },
      milestones: pMilestones,
      team: teamMembers,
      recentComments,
    };
  }, [projects, tasks, milestones, members, comments, projectId]);
}
