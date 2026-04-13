import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi, ApiTask } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';

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
  return useQuery({
    queryKey: ['tasks', effectiveParams],
    queryFn: () => unwrap(flowApi.getTasks(effectiveParams)),
    enabled: !!(effectiveParams.projectId || effectiveParams.workspaceId),
  });
}

export function useMyTasks() {
  return useQuery({
    queryKey: ['my-tasks'],
    queryFn: () => unwrap(flowApi.getMyTasks()),
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
  });
}

export function useLogTime() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ taskId, minutes, note }: { taskId: string; minutes: number; note?: string }) => {
      return unwrap(flowApi.logTime(taskId, { durationMinutes: minutes, description: note })).then(
        () => {}
      );
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['task'] });
    },
  });
}
