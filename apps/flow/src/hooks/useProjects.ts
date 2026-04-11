import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';
import { isFlowLocal } from '../lib/flowLocalMode';
import { useFlowStore } from '../stores/flowStore';

export function useProjects(workspaceId: string | undefined) {
  const projects = useFlowStore((s) => s.projects);
  const localList = useMemo(
    () => (workspaceId ? projects.filter((p) => p.workspaceId === workspaceId) : []),
    [projects, workspaceId]
  );
  const q = useQuery({
    queryKey: ['projects', workspaceId],
    queryFn: () => unwrap(flowApi.getProjects(workspaceId!)),
    enabled: !isFlowLocal() && !!workspaceId,
  });
  return {
    ...q,
    data: isFlowLocal() ? localList : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspaceId: string; name: string; color?: string }) => {
      if (isFlowLocal()) {
        const id = `p-${Date.now()}`;
        const row = {
          id,
          workspaceId: data.workspaceId,
          name: data.name,
          color: data.color ?? '#64C5B9',
          taskCount: 0,
          completedCount: 0,
          memberAvatars: [] as string[],
          health: 'on_track' as const,
          progress: 0,
          startDate: new Date().toISOString().slice(0, 10),
          targetDate: new Date().toISOString().slice(0, 10),
          lead: '',
        };
        useFlowStore.setState({ projects: [...useFlowStore.getState().projects, row] });
        return Promise.resolve(row as any);
      }
      return flowApi.createProject(data);
    },
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['projects', variables.workspaceId] });
    },
  });
}

export function useUpdateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Partial<{ name: string; color: string }>;
    }) => {
      if (isFlowLocal()) {
        const projects = useFlowStore.getState().projects.map((p) =>
          p.id === id ? { ...p, ...data } : p
        );
        useFlowStore.setState({ projects });
        return Promise.resolve();
      }
      return flowApi.updateProject(id, data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  });
}

export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      if (isFlowLocal()) {
        useFlowStore.setState({
          projects: useFlowStore.getState().projects.filter((p) => p.id !== id),
        });
        return Promise.resolve();
      }
      return flowApi.deleteProject(id);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  });
}
