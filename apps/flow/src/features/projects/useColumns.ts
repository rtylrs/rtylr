import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
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
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspaceId: string; name: string; color?: string; sortOrder?: number }) =>
      unwrap(
        flowApi.createProjectColumn({
          project_id: projectId,
          workspace_id: data.workspaceId,
          name: data.name,
          color: data.color,
          sort_order: data.sortOrder,
        })
      ).then((r) => r.column),
    onSuccess: () => qc.invalidateQueries({ queryKey: key(projectId) }),
    onError: () => toast.error('Failed to create column'),
  });
}

export function useUpdateColumn(projectId: string) {
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
    onError: () => toast.error('Failed to update column'),
  });
}

export function useDeleteColumn(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.deleteProjectColumn(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: key(projectId) });
      qc.invalidateQueries({ queryKey: ['tasks'] });
    },
    onError: () => toast.error('Failed to delete column'),
  });
}

export function useMoveTaskToColumn() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, columnId }: { taskId: string; columnId: string }) =>
      unwrap(flowApi.moveTaskToColumn(taskId, columnId)),
    onSuccess: (_data, { taskId }) => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['task', taskId] });
    },
    onError: () => toast.error('Failed to move task'),
  });
}
