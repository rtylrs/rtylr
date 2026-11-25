import { unwrap } from '@voxire/api';
import { dueState } from '../../shared/lib/dateUtils';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiMilestone, ApiStatus, ApiTask } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { fetchAllTasks } from '../tasks/useTasks';
import { taskBucket } from './backlog/backlogModel';

export interface ProjectTaskStats {
  total: number;
  done: number;
  inProgress: number;
  overdue: number;
  progress: number;
}

/**
 * Project progress, counted as the server counts it for the project list and
 * dashboard: top-level tasks that are not archived, with cancelled ones left
 * out of the total and "done" read from the task's status bucket.
 */
export function projectTaskStats(
  tasks: readonly ApiTask[],
  statuses: readonly ApiStatus[]
): ProjectTaskStats {
  const statusesById = new Map(statuses.map((s): [string, ApiStatus] => [s.id, s]));
  let total = 0;
  let done = 0;
  let inProgress = 0;
  let overdue = 0;
  for (const raw of tasks) {
    if (raw.archived || raw.parentTaskId) continue;
    const bucket = taskBucket(apiTaskToTask(raw), statusesById);
    if (bucket === 'cancelled') continue;
    total += 1;
    if (bucket === 'done') done += 1;
    else if (bucket === 'active') inProgress += 1;
    if (bucket !== 'done' && dueState(raw.dueDate) === 'late') overdue += 1;
  }
  return {
    total,
    done,
    inProgress,
    overdue,
    progress: total ? Math.round((done / total) * 100) : 0,
  };
}

export function useProjectOverview(projectId: string | undefined) {
  return useQuery({
    queryKey: ['project-overview', projectId],
    queryFn: async () => {
      if (!projectId) return null;
      const [taskList, statusesRes, msRes] = await Promise.all([
        fetchAllTasks({ projectId }),
        unwrap(flowApi.listStatuses(projectId)),
        unwrap(flowApi.getMilestones(projectId)),
      ]);
      return {
        tasks: taskList,
        stats: projectTaskStats(taskList, statusesRes.statuses ?? []),
        milestones: unwrapListResponse<ApiMilestone>(msRes, 'milestones'),
      };
    },
    enabled: !!projectId,
  });
}
