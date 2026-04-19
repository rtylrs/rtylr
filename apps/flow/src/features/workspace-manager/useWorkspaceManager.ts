import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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
    mutationFn: (data: { workspace_id: number; user_id: number; role?: string }) =>
      unwrap(flowApi.addWorkspaceMember(data)).then(() => {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
  });
}

export function useRemoveWorkspaceMember(workspaceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number }) =>
      unwrap(flowApi.removeWorkspaceMember(data)).then(() => {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
  });
}

export function useUpdateWorkspaceMemberRole(workspaceId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number; role: string }) =>
      unwrap(flowApi.updateWorkspaceMemberRole(data)).then(() => {}),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
  });
}
