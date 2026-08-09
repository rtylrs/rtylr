import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiWorkspace } from '../../shared/services/flowApi';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useWorkspaces() {
  return useQuery<ApiWorkspace[]>({
    queryKey: ['workspaces'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getWorkspaces());
      return unwrapListResponse<ApiWorkspace>(res, 'workspaces');
    },
  });
}

export function useCreateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color: string }) => {
      return unwrap(flowApi.createWorkspace(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
    onError: () => toast.error('Failed to create workspace'),
  });
}

export function useUpdateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<{ name: string; color: string }> }) => {
      return unwrap(flowApi.updateWorkspace(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
    onError: () => toast.error('Failed to update workspace'),
  });
}

export function useDeleteWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      if (id === PERSONAL_WORKSPACE_ID) {
        return Promise.reject(new Error('Personal workspace cannot be deleted'));
      }
      return unwrap(flowApi.deleteWorkspace(id));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
    onError: () => toast.error('Failed to delete workspace'),
  });
}
