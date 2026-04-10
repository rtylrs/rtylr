import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi, ApiMilestone } from '../services/flowApi';

export function useMilestones(projectId?: string) {
  return useQuery({
    queryKey: ['milestones', projectId],
    queryFn: () => flowApi.getMilestones(projectId),
  });
}

export function useCreateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiMilestone>) => flowApi.createMilestone(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
  });
}

export function useUpdateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiMilestone> }) =>
      flowApi.updateMilestone(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
  });
}
