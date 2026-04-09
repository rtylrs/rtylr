import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useTask(id: string | undefined) {
  const { tasks, comments, activity, subtasks, attachments, timelogs, projects } = useFlowStore();
  return useMemo(() => {
    if (!id) return null;
    const task = tasks.find((t) => t.id === id);
    if (!task) return null;
    return {
      task,
      project: projects.find((p) => p.id === task.projectId),
      comments: comments.filter((c) => c.taskId === id).sort((a, b) => a.createdAt.localeCompare(b.createdAt)),
      activity: activity.filter((a) => a.taskId === id).sort((a, b) => b.createdAt.localeCompare(a.createdAt)),
      subtasks: subtasks.filter((s) => s.taskId === id),
      attachments: attachments.filter((a) => a.taskId === id),
      timelogs: timelogs.filter((t) => t.taskId === id),
    };
  }, [id, tasks, comments, activity, subtasks, attachments, timelogs, projects]);
}
