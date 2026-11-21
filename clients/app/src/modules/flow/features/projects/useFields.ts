import { getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import {
  flowApi,
  type ApiField,
  type ApiFieldInput,
  type ApiFieldType,
  type ApiTaskFieldValueInput,
} from '../../shared/services/flowApi';

function key(projectId: string) {
  return ['fields', projectId];
}

// A field change reshapes every task in the project, so the task caches go too.
function invalidateFieldViews(qc: QueryClient, projectId: string): void {
  qc.invalidateQueries({ queryKey: key(projectId) });
  qc.invalidateQueries({ queryKey: ['tasks'] });
  qc.invalidateQueries({ queryKey: ['task'] });
}

/** A project's custom fields in their display order. */
export function useProjectFields(projectId: string) {
  return useQuery<ApiField[]>({
    queryKey: key(projectId),
    queryFn: async () => {
      const res = await unwrap(flowApi.listFields(projectId));
      return res.fields ?? [];
    },
    enabled: !!projectId,
  });
}

export function useCreateField(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ type, ...data }: ApiFieldInput & { type: ApiFieldType }) =>
      unwrap(flowApi.createField(projectId, type, data)).then((r) => r.field),
    onSuccess: () => {
      invalidateFieldViews(qc, projectId);
      toast.success(t('flowToasts.field.created'));
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('flowToasts.field.saveFailed'))),
  });
}

export function useUpdateField(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: ApiFieldInput & { id: string }) =>
      unwrap(flowApi.updateField(id, data)).then((r) => r.field),
    onSuccess: () => {
      invalidateFieldViews(qc, projectId);
      toast.success(t('flowToasts.field.updated'));
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, t('flowToasts.field.saveFailed'))),
  });
}

/** Sets the whole order at once; the list moves straight away and snaps back on failure. */
export function useReorderFields(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (ids: string[]) =>
      unwrap(flowApi.reorderFields(projectId, ids)).then((r) => r.fields ?? []),
    onMutate: async (ids) => {
      await qc.cancelQueries({ queryKey: key(projectId) });
      const previous = qc.getQueryData<ApiField[]>(key(projectId));
      if (previous) {
        const byId = new Map(previous.map((f) => [f.id, f]));
        qc.setQueryData(
          key(projectId),
          ids.map((id) => byId.get(id)).filter((f): f is ApiField => Boolean(f))
        );
      }
      return { previous };
    },
    onError: (_err, _ids, context) => {
      if (context?.previous) qc.setQueryData(key(projectId), context.previous);
      toast.error(t('flowToasts.field.saveFailed'));
    },
    onSettled: () => invalidateFieldViews(qc, projectId),
  });
}

export function useDeleteField(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.deleteField(id)),
    onSuccess: (res) => {
      invalidateFieldViews(qc, projectId);
      toast.success(t('flowToasts.field.deleted', { count: res.clearedTaskCount }));
    },
    onError: () => toast.error(t('flowToasts.field.deleteFailed')),
  });
}

/** Writes one task's values; the server answers with the full set, which goes straight into the cache. */
export function useSetTaskFieldValues(taskId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (values: ApiTaskFieldValueInput[]) =>
      unwrap(flowApi.setTaskFieldValues(taskId, values)).then((r) => r.fieldValues),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task', taskId] });
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: (error) =>
      toast.error(getOneKitErrorMessage(error, t('flowToasts.field.valueSaveFailed'))),
  });
}
