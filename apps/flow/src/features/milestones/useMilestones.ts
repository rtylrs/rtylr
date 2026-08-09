import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiMilestone } from '../../shared/services/flowApi';

export function useMilestones(projectId?: string) {
  return useQuery<ApiMilestone[]>({
    queryKey: ['milestones', projectId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getMilestones(projectId));
      if (Array.isArray(res)) return res;
      const wrapped = res as { milestones?: ApiMilestone[] };
      return wrapped.milestones ?? [];
    },
  });
}

export function useCreateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiMilestone>) => {
      return unwrap(flowApi.createMilestone(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
    onError: () => toast.error('Failed to create milestone'),
  });
}

export function useUpdateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiMilestone> }) => {
      return unwrap(flowApi.updateMilestone(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
    onError: () => toast.error('Failed to update milestone'),
  });
}
