import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiProjectMember } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useProjectMembers(projectId?: string) {
  return useQuery<ApiProjectMember[]>({
    queryKey: ['project-members', projectId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getProjectMembers(projectId!));
      return unwrapListResponse<ApiProjectMember>(res, 'members');
    },
    enabled: !!projectId,
  });
}

export function useAddProjectMember(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; user_id: number; role?: string }) =>
      unwrap(flowApi.addProjectMember(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['project-members', projectId] }),
    onError: () => toast.error('Failed to add member'),
  });
}

export function useRemoveProjectMember(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; user_id: number }) =>
      unwrap(flowApi.removeProjectMember(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['project-members', projectId] }),
    onError: () => toast.error('Failed to remove member'),
  });
}

export function useUpdateProjectMemberRole(projectId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; user_id: number; role: string }) =>
      unwrap(flowApi.updateProjectMemberRole(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['project-members', projectId] }),
    onError: () => toast.error('Failed to update role'),
  });
}

export function useTransferProjectOwnership() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; new_owner_id: number }) =>
      unwrap(flowApi.transferProjectOwnership(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error('Failed to transfer ownership'),
  });
}

export function useInviteProjectMember(projectId: string) {
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
      toast.success('Invitation sent');
    },
    onError: () => toast.error('No user found with that email'),
  });
}
