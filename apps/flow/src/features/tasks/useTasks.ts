import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi, ApiTask } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { isTaskMine } from '../../shared/lib/currentUser';
import { useFlowStore, type Task } from '../../shared/stores/flowStore';

export function useTasks(params: {
  projectId?: string;
  status?: ApiTask['status'];
  assigneeId?: string;
}) {
  const tasks = useFlowStore((s) => s.tasks);
  const localList = useMemo(() => {
    let list = tasks.filter((t) => !t.archived);
    if (params.projectId) list = list.filter((t) => t.projectId === params.projectId);
    if (params.status) list = list.filter((t) => t.status === params.status);
    if (params.assigneeId) list = list.filter((t) => t.assigneeId === params.assigneeId);
    return list;
  }, [tasks, params.projectId, params.status, params.assigneeId]);

  const q = useQuery({
    queryKey: ['tasks', params],
    queryFn: () => unwrap(flowApi.getTasks(params)),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? localList : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}

export function useMyTasks() {
  const tasks = useFlowStore((s) => s.tasks);
  const localList = useMemo(
    () => tasks.filter((t) => isTaskMine(t.assigneeId) && !t.archived),
    [tasks]
  );
  const q = useQuery({
    queryKey: ['my-tasks'],
    queryFn: () => unwrap(flowApi.getMyTasks()),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? localList : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}

export function useTaskDetail(id: string | undefined) {
  const fromStore = useFlowStore((s) => (id ? s.tasks.find((t) => t.id === id) : undefined));
  const q = useQuery({
    queryKey: ['task', id],
    queryFn: () => unwrap(flowApi.getTaskDetail(id!)),
    enabled: !isFlowLocal() && !!id,
  });
  return {
    ...q,
    data: isFlowLocal() ? fromStore : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
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
      if (isFlowLocal()) {
        const created = useFlowStore.getState().createTask({
          title: data.title,
          projectId: data.projectId,
          workspaceId: data.workspaceId,
          status: data.status as Task['status'] | undefined,
          priority: data.priority,
          assigneeId: data.assigneeId,
          dueDate: data.dueDate,
        });
        return Promise.resolve(created as unknown as ApiTask);
      }
      return flowApi.createTask(data);
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
      if (isFlowLocal()) {
        useFlowStore.getState().updateTask(id, data as Partial<Task>);
        return Promise.resolve();
      }
      return flowApi.updateTask(id, data);
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
      if (isFlowLocal()) {
        useFlowStore.getState().deleteTask(id);
        return Promise.resolve();
      }
      return flowApi.deleteTask(id);
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
      if (isFlowLocal()) {
        useFlowStore.getState().updateTask(id, { status: status as Task['status'] });
        return Promise.resolve();
      }
      return flowApi.updateTaskStatus(id, status);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tasks'] });
      qc.invalidateQueries({ queryKey: ['my-tasks'] });
    },
  });
}
