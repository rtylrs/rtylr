import { toFiniteNumber, unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { flowApi, ApiWorkspaceMember } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useWorkspaceMembers(workspaceId?: string) {
  return useQuery<ApiWorkspaceMember[]>({
    queryKey: ['workspace-members', workspaceId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getWorkspaceMembers(workspaceId!));
      return unwrapListResponse<ApiWorkspaceMember>(res, 'members');
    },
    enabled: !!workspaceId,
  });
}

export function useAddWorkspaceMember(workspaceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number; role?: string }) => {
      const wsId = toFiniteNumber(data.workspace_id);
      const uid = toFiniteNumber(data.user_id);
      if (wsId === 0) {
        toast.error('Invalid workspace ID');
        return Promise.reject(new Error('Invalid workspace ID'));
      }
      if (uid === 0) {
        toast.error('Invalid user ID');
        return Promise.reject(new Error('Invalid user ID'));
      }
      return unwrap(flowApi.addWorkspaceMember(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
    onError: () => toast.error('Failed to add workspace member'),
  });
}

export function useRemoveWorkspaceMember(workspaceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number }) => {
      const wsId = toFiniteNumber(data.workspace_id);
      const uid = toFiniteNumber(data.user_id);
      if (wsId === 0) {
        toast.error('Invalid workspace ID');
        return Promise.reject(new Error('Invalid workspace ID'));
      }
      if (uid === 0) {
        toast.error('Invalid user ID');
        return Promise.reject(new Error('Invalid user ID'));
      }
      return unwrap(flowApi.removeWorkspaceMember(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
    onError: () => toast.error('Failed to remove workspace member'),
  });
}

export function useUpdateWorkspaceMemberRole(workspaceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number; role: string }) => {
      const wsId = toFiniteNumber(data.workspace_id);
      const uid = toFiniteNumber(data.user_id);
      if (wsId === 0) {
        toast.error('Invalid workspace ID');
        return Promise.reject(new Error('Invalid workspace ID'));
      }
      if (uid === 0) {
        toast.error('Invalid user ID');
        return Promise.reject(new Error('Invalid user ID'));
      }
      return unwrap(flowApi.updateWorkspaceMemberRole(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
    onError: () => toast.error('Failed to update workspace member role'),
  });
}
