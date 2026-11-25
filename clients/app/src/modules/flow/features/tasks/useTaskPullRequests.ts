import { useQuery } from '@tanstack/react-query';
import type { TaskPullRequest } from '@voxire/voxie/flow/github/v2/types';
import { githubClient } from '../../../../shared/lib/clients';

export type { TaskPullRequest };

/** The PRs linked to a task with their CI checks (T-376). */
export function useTaskPullRequests(taskId: string | undefined, enabled = true) {
  return useQuery({
    queryKey: ['task-pull-requests', taskId],
    queryFn: async () =>
      (await githubClient.listTaskPullRequests({ taskId: taskId! })).pullRequests ?? [],
    enabled: enabled && !!taskId,
    staleTime: 30 * 1000,
  });
}
