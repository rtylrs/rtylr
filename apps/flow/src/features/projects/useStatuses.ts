import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi, ApiStatus } from '../../shared/services/flowApi';

function key(projectId: string) {
  return ['statuses', projectId];
}

export function useProjectStatuses(projectId: string) {
  return useQuery<ApiStatus[]>({
    queryKey: key(projectId),
    queryFn: async () => {
      const res = await unwrap(flowApi.listStatuses(projectId));
      return res.statuses ?? [];
    },
    enabled: !!projectId,
  });
}

export function useCreateStatus(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      name: string;
      bucket: string;
      color?: string;
      sortOrder?: number;
      isDefault?: boolean;
    }) => unwrap(flowApi.createStatus({ projectId, ...data })).then((r) => r.status),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(projectId) }),
    onError: () => toast.error(t('toasts.status.createFailed')),
  });
}

export function useUpdateStatus(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      id: string;
      name: string;
      bucket: string;
      color: string;
      sortOrder?: number;
      isDefault?: boolean;
    }) => unwrap(flowApi.updateStatus(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(projectId) }),
    onError: () => toast.error(t('toasts.status.updateFailed')),
  });
}

export function useDeleteStatus(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.deleteStatus(id)),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: key(projectId) });
      qc.invalidateQueries({ queryKey: ['tasks'] });
      if (res.reassignedTaskCount > 0) {
        toast.success(
          t('toasts.status.deletedWithReassignment', { count: res.reassignedTaskCount })
        );
      } else {
        toast.success(t('toasts.status.deleted'));
      }
    },
    onError: () => toast.error(t('toasts.status.deleteFailed')),
  });
}

export function useSetTaskStatus() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { taskId: string; statusId: string }) => unwrap(flowApi.setTaskStatus(data)),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['task', variables.taskId] });
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('toasts.status.setFailed')),
  });
}
