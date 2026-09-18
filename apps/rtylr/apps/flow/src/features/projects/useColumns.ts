import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { unwrap } from '@voxire/api';
import { flowApi, ApiProjectColumn } from '../../shared/services/flowApi';

function key(projectId: string) {
  return ['columns', projectId];
}

export function useProjectColumns(projectId: string) {
  return useQuery<ApiProjectColumn[]>({
    queryKey: key(projectId),
    queryFn: async () => {
      const res = await unwrap(flowApi.getProjectColumns(projectId));
      return res.columns ?? [];
    },
    enabled: !!projectId,
  });
}

export function useCreateColumn(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color?: string; sortOrder?: number }) =>
      unwrap(
        flowApi.createProjectColumn({
          project_id: projectId,
          name: data.name,
          color: data.color,
          sort_order: data.sortOrder,
        })
      ).then((r) => r.column),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(projectId) }),
    onError: () => toast.error(t('toasts.column.createFailed')),
  });
}

export function useUpdateColumn(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { id: string; name: string; color?: string; sortOrder?: number }) =>
      unwrap(
        flowApi.updateProjectColumn({
          id: data.id,
          name: data.name,
          color: data.color,
          sort_order: data.sortOrder,
        })
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(projectId) }),
    onError: () => toast.error(t('toasts.column.updateFailed')),
  });
}

export function useDeleteColumn(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.deleteProjectColumn(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: key(projectId) });
      qc.invalidateQueries({ queryKey: ['tasks'] });
    },
    onError: () => toast.error(t('toasts.column.deleteFailed')),
  });
}

export function useMoveTaskToColumn() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, columnId }: { taskId: string; columnId: string }) =>
      unwrap(flowApi.moveTaskToColumn(taskId, columnId)),
    onSuccess: (_data, { taskId }) => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['task', taskId] });
    },
    onError: () => toast.error(t('toasts.task.moveFailed')),
  });
}
