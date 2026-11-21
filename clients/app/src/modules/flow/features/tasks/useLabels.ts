import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiLabel } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

/** `enabled: false` holds the request until a caller actually needs the labels. */
export function useLabels({ enabled = true }: { enabled?: boolean } = {}) {
  return useQuery<ApiLabel[]>({
    queryKey: ['labels'],
    queryFn: async () => {
      const res = await unwrap(flowApi.listLabels());
      return unwrapListResponse<ApiLabel>(res, 'labels');
    },
    enabled,
  });
}

export function useCreateLabel() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color: string }) => unwrap(flowApi.createLabel(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['labels'] }),
    onError: () => toast.error(t('flowToasts.label.createFailed')),
  });
}

export function useDeleteLabel() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(flowApi.deleteLabel(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['labels'] }),
    onError: () => toast.error(t('flowToasts.label.deleteFailed')),
  });
}

export function useSetTaskLabels() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { task_id: number; label_ids: number[] }) =>
      unwrap(flowApi.setTaskLabels(data)),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['task', String(variables.task_id)] });
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('flowToasts.label.setFailed')),
  });
}
