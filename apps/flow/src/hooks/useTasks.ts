import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi, ApiTask } from '../services/flowApi';

export function useTasks(params: { projectId?: string; status?: ApiTask['status']; assigneeId?: string }) {
  return useQuery({
    queryKey: ['tasks', params],
    queryFn: () => flowApi.getTasks(params),
  });
}

export function useMyTasks() {
  return useQuery({
    queryKey: ['my-tasks'],
    queryFn: () => flowApi.getMyTasks(),
  });
}

export function useTaskDetail(id: string | undefined) {
  return useQuery({
    queryKey: ['task', id],
    queryFn: () => flowApi.getTaskDetail(id!),
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
    }) => flowApi.createTask(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
  });
}

export function useUpdateTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiTask> }) => flowApi.updateTask(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
  });
}

export function useDeleteTask() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => flowApi.deleteTask(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
  });
}

export function useUpdateTaskStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: ApiTask['status'] }) =>
      flowApi.updateTaskStatus(id, status),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
  });
}
