import { toFiniteNumber, unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
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
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number; role?: string }) => {
      const wsId = toFiniteNumber(data.workspace_id);
      const uid = toFiniteNumber(data.user_id);
      if (wsId === 0) {
        toast.error(t('toasts.workspaceMember.invalidWorkspaceId'));
        return Promise.reject(new Error('Invalid workspace ID'));
      }
      if (uid === 0) {
        toast.error(t('toasts.workspaceMember.invalidUserId'));
        return Promise.reject(new Error('Invalid user ID'));
      }
      return unwrap(flowApi.addWorkspaceMember(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
    onError: () => toast.error(t('toasts.workspaceMember.addWorkspaceMemberFailed')),
  });
}

export function useRemoveWorkspaceMember(workspaceId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number }) => {
      const wsId = toFiniteNumber(data.workspace_id);
      const uid = toFiniteNumber(data.user_id);
      if (wsId === 0) {
        toast.error(t('toasts.workspaceMember.invalidWorkspaceId'));
        return Promise.reject(new Error('Invalid workspace ID'));
      }
      if (uid === 0) {
        toast.error(t('toasts.workspaceMember.invalidUserId'));
        return Promise.reject(new Error('Invalid user ID'));
      }
      return unwrap(flowApi.removeWorkspaceMember(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
    onError: () => toast.error(t('toasts.workspaceMember.removeWorkspaceMemberFailed')),
  });
}

export function useUpdateWorkspaceMemberRole(workspaceId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspace_id: number; user_id: number; role: string }) => {
      const wsId = toFiniteNumber(data.workspace_id);
      const uid = toFiniteNumber(data.user_id);
      if (wsId === 0) {
        toast.error(t('toasts.workspaceMember.invalidWorkspaceId'));
        return Promise.reject(new Error('Invalid workspace ID'));
      }
      if (uid === 0) {
        toast.error(t('toasts.workspaceMember.invalidUserId'));
        return Promise.reject(new Error('Invalid user ID'));
      }
      return unwrap(flowApi.updateWorkspaceMemberRole(data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] });
    },
    onError: () => toast.error(t('toasts.workspaceMember.updateRoleFailed')),
  });
}
