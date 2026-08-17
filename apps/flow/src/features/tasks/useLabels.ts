import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiLabel } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useLabels(workspaceId?: string) {
  return useQuery<ApiLabel[]>({
    queryKey: ['labels', workspaceId],
    queryFn: async () => {
      if (workspaceId) {
        const res = await unwrap(flowApi.listLabels(workspaceId));
        return unwrapListResponse<ApiLabel>(res, 'labels');
      }
      const res = await unwrap(flowApi.getLabels());
      return unwrapListResponse<ApiLabel>(res, 'labels');
    },
  });
}

export function useCreateLabel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: string; name: string; color: string }) =>
      unwrap(flowApi.createLabel(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['labels'] }),
    onError: () => toast.error('Failed to create label'),
  });
}

export function useUpdateLabel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: number; name: string; color: string }) =>
      unwrap(flowApi.updateLabel(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['labels'] }),
    onError: () => toast.error('Failed to update label'),
  });
}

export function useDeleteLabel() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(flowApi.deleteLabel(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['labels'] }),
    onError: () => toast.error('Failed to delete label'),
  });
}

export function useSetTaskLabels() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { task_id: number; label_ids: number[] }) =>
      unwrap(flowApi.setTaskLabels(data)),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['task', String(variables.task_id)] });
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error('Failed to set task labels'),
  });
}
