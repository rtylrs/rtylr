import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ProjectServiceClient } from '@voxire/voxie/flow/project/v2/client';
import type { ProjectLabourCost } from '@voxire/voxie/flow/project/v2/types';

// The project service serves the labour-cost aggregate via the HR-owned
// labourcost port. Rates never cross the boundary; only aggregate cents and
// coverage counts.
const projectClient = createOneKitClient(ProjectServiceClient, 'flow');

export const LABOUR_COST_KEY = (projectId: string) =>
  ['flow', 'project', 'labour-cost', projectId] as const;

/**
 * Live labour cost for one Flow project, computed server-side by HR.
 * Admin-gated: only users with the admin project role can see it.
 *
 * Degrade (#837 / rtylr#837): a tenant without HR — or a failed read —
 * yields null and the panel hides; never a zero-cost claim.
 */
export function useProjectLabourCost(projectId: string | undefined, enabled = true) {
  return useQuery<ProjectLabourCost | null>({
    queryKey: LABOUR_COST_KEY(projectId ?? ''),
    queryFn: async (): Promise<ProjectLabourCost | null> => {
      if (!projectId) return null;
      try {
        const res = await projectClient.getProjectLabourCost({ id: projectId });
        return res.cost ?? null;
      } catch {
        return null;
      }
    },
    enabled: enabled && !!projectId,
    staleTime: 2 * 60 * 1000,
    retry: false,
  });
}
