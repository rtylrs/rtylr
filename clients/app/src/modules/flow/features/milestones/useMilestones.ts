import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiMilestone } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useMilestones(projectId?: string) {
  return useQuery<ApiMilestone[]>({
    queryKey: projectId === undefined ? ['milestones'] : ['milestones', projectId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getMilestones(projectId));
      return unwrapListResponse<ApiMilestone>(res, 'milestones');
    },
    enabled: projectId !== '',
  });
}

export function useCreateMilestone() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiMilestone> & { projectId: string }) => {
      return unwrap(flowApi.createMilestone(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
    onError: () => toast.error(t('flowToasts.milestone.createFailed')),
  });
}

export function useUpdateMilestone() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiMilestone> }) => {
      return unwrap(flowApi.updateMilestone(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
    onError: () => toast.error(t('flowToasts.milestone.updateFailed')),
  });
}
