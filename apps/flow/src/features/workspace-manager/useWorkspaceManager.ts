import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi, ApiWorkspaceMember } from '../../shared/services/flowApi';

export function useWorkspaceMembers(workspaceId?: string) {
  return useQuery({
    queryKey: ['workspace-members', workspaceId],
    queryFn: () => unwrap(flowApi.getWorkspaceMembers(workspaceId!)),
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
