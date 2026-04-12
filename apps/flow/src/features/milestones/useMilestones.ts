import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiMilestone } from '../../shared/services/flowApi';

export function useMilestones(projectId?: string) {
  return useQuery({
    queryKey: ['milestones', projectId],
    queryFn: () => unwrap(flowApi.getMilestones(projectId)),
  });
}

export function useCreateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiMilestone>) => {
      return unwrap(flowApi.createMilestone(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
  });
}

export function useUpdateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiMilestone> }) => {
      return unwrap(flowApi.updateMilestone(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
  });
}
