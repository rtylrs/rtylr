import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { useFlowStore } from '../../shared/stores/flowStore';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';

export function useWorkspaces() {
  const local = useFlowStore((s) => s.workspaces);
  const q = useQuery({
    queryKey: ['workspaces'],
    queryFn: () => unwrap(flowApi.getWorkspaces()),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? local : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}

export function useCreateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color: string }) => {
      if (isFlowLocal()) {
        const id = `ws-${Date.now()}`;
        useFlowStore.getState().setWorkspaces([
          ...useFlowStore.getState().workspaces,
          { id, name: data.name, color: data.color },
        ]);
        return Promise.resolve({ id, name: data.name, color: data.color });
      }
      return flowApi.createWorkspace(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
  });
}

export function useUpdateWorkspace() {
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
        const ws = useFlowStore.getState().workspaces.map((w) =>
          w.id === id ? { ...w, ...data } : w
        );
        useFlowStore.setState({ workspaces: ws });
        return Promise.resolve();
      }
      return flowApi.updateWorkspace(id, data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
  });
}

export function useDeleteWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      if (id === PERSONAL_WORKSPACE_ID) {
        return Promise.reject(new Error('Personal workspace cannot be deleted'));
      }
      if (isFlowLocal()) {
        const state = useFlowStore.getState();
        const next = state.workspaces.filter((w) => w.id !== id);
        const nextActive =
          state.activeWorkspaceId === id
            ? (next.find((w) => w.id === PERSONAL_WORKSPACE_ID)?.id ?? next[0]?.id ?? '')
            : state.activeWorkspaceId;
        useFlowStore.setState({ workspaces: next, activeWorkspaceId: nextActive });
        return Promise.resolve();
      }
      return flowApi.deleteWorkspace(id);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
  });
}
