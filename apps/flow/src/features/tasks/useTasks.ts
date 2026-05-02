import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';
import { flowApi, ApiTask } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

const t = (key: string, opts?: Record<string, unknown>): string => i18next.t(key, opts) as string;

export function useTasks(params: {
  projectId?: string;
  status?: ApiTask['status'];
  assigneeId?: string;
}) {
  const { activeWorkspaceId } = useFlowStore();
  const effectiveParams = {
    ...params,
    workspaceId: params.projectId ? undefined : activeWorkspaceId || undefined,
  };
  return useQuery<ApiTask[]>({
    queryKey: ['tasks', effectiveParams],
    queryFn: async () => {
      const res = await unwrap(flowApi.getTasks(effectiveParams));
      return unwrapListResponse<ApiTask>(res, 'tasks');
    },
    enabled: !!(effectiveParams.projectId || effectiveParams.workspaceId),
  });
}

export function useMyTasks() {
  return useQuery<ApiTask[]>({
    queryKey: ['my-tasks'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getMyTasks());
      return unwrapListResponse<ApiTask>(res, 'tasks');
    },
  });
}

export function useTaskDetail(id: string | undefined) {
  return useQuery({
    queryKey: ['task', id],
    queryFn: () => unwrap(flowApi.getTaskDetail(id!)),
    enabled: !!id,
  });
}

export function useCreateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      projectId: string;
      workspaceId: string;
      title: string;
      status?: ApiTask['status'];
      priority?: ApiTask['priority'];
      assigneeId?: string;
      dueDate?: string;
    }) => {
      return unwrap(flowApi.createTask(data));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('toasts.task.createFailed')),
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiTask> }) => {
      return unwrap(flowApi.updateTask(id, data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.task.updateFailed')),
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteTask(id));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('toasts.task.deleteFailed')),
  });
}

export function useUpdateTaskStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: ApiTask['status'] }) => {
      return unwrap(flowApi.updateTaskStatus(id, status));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
    onError: () => toast.error(t('toasts.task.statusUpdateFailed')),
  });
}

export function useBulkUpdateStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ ids, status }: { ids: string[]; status: ApiTask['status'] }) => {
      await Promise.all(ids.map((id) => unwrap(flowApi.updateTaskStatus(id, status))));
    },
    onSuccess: (_data, { ids }) => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
      toast.success(t('toasts.task.bulkUpdated', { count: ids.length }));
    },
    onError: () => toast.error(t('toasts.task.bulkUpdateFailed')),
  });
}

export function useBulkUpdatePriority() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ ids, priority }: { ids: string[]; priority: ApiTask['priority'] }) => {
      await Promise.all(ids.map((id) => unwrap(flowApi.updateTask(id, { priority }))));
    },
    onSuccess: (_data, { ids }) => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
      toast.success(t('toasts.task.bulkUpdated', { count: ids.length }));
    },
    onError: () => toast.error(t('toasts.task.bulkUpdateFailed')),
  });
}

export function useBulkDelete() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (ids: string[]) => {
      await Promise.all(ids.map((id) => unwrap(flowApi.deleteTask(id))));
    },
    onSuccess: (_data, ids) => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
      toast.success(t('toasts.task.bulkDeleted', { count: ids.length }));
    },
    onError: () => toast.error(t('toasts.task.bulkDeleteFailed')),
  });
}

export function useAddComment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { taskId: string; content: string }) => {
      return unwrap(flowApi.addComment(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.comment.postFailed')),
  });
}

export function useDeleteComment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteComment(id));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.comment.deleteFailed')),
  });
}

export function useAddSubtask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, title }: { taskId: string; title: string }) => {
      return unwrap(flowApi.createSubtask(taskId, { title })).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.subtask.addFailed')),
  });
}

export function useToggleSubtask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, completed }: { id: string; completed: boolean }) =>
      unwrap(flowApi.updateTask(id, { status: completed ? 'todo' : 'done' })).then(() => {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.subtask.updateFailed')),
  });
}

export function useLogTime() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      task_id,
      duration_minutes,
      description,
    }: {
      task_id: number;
      duration_minutes: number;
      description?: string;
    }) => {
      return unwrap(
        flowApi.logTime({
          task_id,
          duration_minutes,
          description,
        })
      ).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
    onError: () => toast.error(t('toasts.time.logFailed')),
  });
}
