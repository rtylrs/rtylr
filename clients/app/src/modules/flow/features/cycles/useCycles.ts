import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiCycle } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

/**
 * The query behind useCycles, shared so a screen that needs several projects'
 * cycles at once (the Plan calendar) reads and fills the same cache entries.
 */
export function cyclesQueryOptions(projectId?: string) {
  return {
    queryKey: ['cycles', projectId],
    queryFn: async (): Promise<ApiCycle[]> => {
      const res = await unwrap(flowApi.getCycles(projectId!));
      return unwrapListResponse<ApiCycle>(res, 'cycles');
    },
    enabled: projectId !== undefined && projectId !== '',
  };
}

/**
 * Cycles for one project.
 *
 * v2 requires a project id, so this stays disabled until the caller has one
 * rather than firing a request that would 400.
 */
export function useCycles(projectId?: string): ReturnType<typeof useQuery<ApiCycle[]>> {
  return useQuery<ApiCycle[]>(cyclesQueryOptions(projectId));
}

export function useCreateCycle() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiCycle> & { projectId: string }) => {
      return unwrap(flowApi.createCycle(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
    onError: () => toast.error(t('flowToasts.cycle.createFailed')),
  });
}

export function useUpdateCycle() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiCycle> }) => {
      return unwrap(flowApi.updateCycle(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
    onError: () => toast.error(t('flowToasts.cycle.updateFailed')),
  });
}
