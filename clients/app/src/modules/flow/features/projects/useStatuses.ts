import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi, ApiStatus, type ApiStatusTemplateItem } from '../../shared/services/flowApi';
import type { SortUpdate } from './statusUtils';

function key(projectId: string) {
  return ['statuses', projectId];
}

const TEMPLATE_KEY = ['status-template'];

function invalidateStatusViews(qc: QueryClient, projectId: string): void {
  qc.invalidateQueries({ queryKey: key(projectId) });
  qc.invalidateQueries({ queryKey: ['tasks'] });
  qc.invalidateQueries({ queryKey: ['my-tasks'] });
  qc.invalidateQueries({ queryKey: ['task'] });
}

function writeSortOrder({ status, sortOrder }: SortUpdate) {
  return unwrap(
    flowApi.updateStatus({
      id: status.id,
      name: status.name,
      bucket: status.bucket,
      color: status.color,
      sortOrder,
      isDefault: status.isDefault,
    })
  );
}

/** A project's statuses, which are also its board columns. */
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

export interface CreateStatusInput {
  name: string;
  bucket: string;
  color?: string;
  sortOrder?: number;
  isDefault?: boolean;
  /** Sort-order writes that make room for the new status first. */
  makeRoom?: SortUpdate[];
}

export function useCreateStatus(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ makeRoom = [], ...data }: CreateStatusInput) => {
      await Promise.all(makeRoom.map(writeSortOrder));
      return unwrap(flowApi.createStatus({ projectId, ...data })).then((r) => r.status);
    },
    onSuccess: () => invalidateStatusViews(qc, projectId),
    onError: () => {
      qc.invalidateQueries({ queryKey: key(projectId) });
      toast.error(t('flowToasts.status.createFailed'));
    },
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
    onSuccess: () => invalidateStatusViews(qc, projectId),
    onError: () => toast.error(t('flowToasts.status.updateFailed')),
  });
}

/**
 * Applies a reorder as independent writes: there is no bulk endpoint, and if
 * one fails the refetch shows the server's real order rather than a guess.
 */
export function useReorderStatuses(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (updates: SortUpdate[]) => Promise.all(updates.map(writeSortOrder)),
    onSettled: () => invalidateStatusViews(qc, projectId),
    onError: () => toast.error(t('flowToasts.status.updateFailed')),
  });
}

export function useDeleteStatus(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, moveToStatusId }: { id: string; moveToStatusId?: string }) =>
      unwrap(flowApi.deleteStatus(id, moveToStatusId)),
    onSuccess: (res) => {
      invalidateStatusViews(qc, projectId);
      if (res.reassignedTaskCount > 0) {
        toast.success(t('flowToasts.status.deletedWithMove', { count: res.reassignedTaskCount }));
      } else {
        toast.success(t('flowToasts.status.deleted'));
      }
    },
    onError: () => toast.error(t('flowToasts.status.deleteFailed')),
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
    onError: () => toast.error(t('flowToasts.status.setFailed')),
  });
}

/** The workspace template new projects copy their statuses from. */
export function useStatusTemplate() {
  return useQuery<ApiStatusTemplateItem[]>({
    queryKey: TEMPLATE_KEY,
    queryFn: async () => {
      const res = await unwrap(flowApi.getStatusTemplate());
      return res.statuses ?? [];
    },
  });
}

export function useSaveStatusTemplate() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (statuses: ApiStatusTemplateItem[]) =>
      unwrap(flowApi.saveStatusTemplate(statuses)).then((r) => r.statuses ?? []),
    onSuccess: (saved) => {
      qc.setQueryData(TEMPLATE_KEY, saved);
      toast.success(t('flowToasts.status.templateSaved'));
    },
    onError: () => toast.error(t('flowToasts.status.templateSaveFailed')),
  });
}
