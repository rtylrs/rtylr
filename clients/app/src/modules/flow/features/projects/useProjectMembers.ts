import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiProjectMember } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

/**
 * Query options for one project's member list, shared by `useProjectMembers` and
 * by the multi-project lookup in `useCurrentProjectRoles`, so both read and fill
 * the same `['project-members', projectId]` cache entry rather than refetching.
 */
export function projectMembersQueryOptions(projectId?: string) {
  return {
    queryKey: ['project-members', projectId],
    queryFn: async (): Promise<ApiProjectMember[]> => {
      const res = await unwrap(flowApi.getProjectMembers(projectId!));
      return unwrapListResponse<ApiProjectMember>(res, 'members');
    },
    enabled: !!projectId,
  };
}

export function useProjectMembers(projectId?: string) {
  return useQuery<ApiProjectMember[]>(projectMembersQueryOptions(projectId));
}

export function useRemoveProjectMember(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; user_id: number }) =>
      unwrap(flowApi.removeProjectMember(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['project-members', projectId] }),
    onError: () => toast.error(t('flowToasts.projectMember.removeFailed')),
  });
}

export function useUpdateProjectMemberRole(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; user_id: number; role: string }) =>
      unwrap(flowApi.updateProjectMemberRole(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['project-members', projectId] }),
    onError: () => toast.error(t('flowToasts.projectMember.roleUpdateFailed')),
  });
}

export function useTransferProjectOwnership() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; new_owner_id: number }) =>
      unwrap(flowApi.transferProjectOwnership(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error(t('flowToasts.projectMember.transferFailed')),
  });
}

export function useInviteProjectMember(projectId: string) {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { email: string; role: string }) =>
      unwrap(
        flowApi.inviteProjectMember({
          project_id: Number(projectId),
          email: data.email,
          role: data.role,
        })
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['project-members', projectId] });
      toast.success(t('flowToasts.projectMember.invited'));
    },
    onError: () => toast.error(t('flowToasts.projectMember.notFoundByEmail')),
  });
}
