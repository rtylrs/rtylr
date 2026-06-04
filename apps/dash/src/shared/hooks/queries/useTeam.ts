import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { StaffUserPayload } from '@voxire/contracts/hub/team/members/v1/types';
import { z } from 'zod';
import api, { safe } from '../../lib/api';
import { teamMemberSchema } from '../../schemas/team';
import { toast } from 'sonner';

export type { TeamMember } from '../../schemas/team';
type TeamMember = z.infer<typeof teamMemberSchema> & { status: string };

interface UseTeamParams {
  page?: number;
  pageSize?: number;
}

export interface TeamListResult {
  members: TeamMember[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export const useTeam = (params: UseTeamParams = {}) => {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<TeamListResult>({
    queryKey: ['team', { page, pageSize }],
    queryFn: async (): Promise<TeamListResult> => {
      const res = await unwrap(
        safe.get('/team/members', { params: { page, page_size: pageSize } })
      );
      const raw = res as Record<string, unknown>;
      const list = raw.users ?? raw.members ?? res;
      const arr = Array.isArray(list) ? list : [];
      const members = arr.map((m: StaffUserPayload) => {
        const parsed = teamMemberSchema.parse(m);
        return {
          ...parsed,
          name: parsed.fullName || parsed.name || parsed.email,
          status: parsed.isActive ? 'active' : 'inactive',
          groups: parsed.groups ?? [],
        } as TeamMember;
      });
      return {
        members,
        hasMore: Boolean(raw.has_more),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.page_size ?? pageSize),
      };
    },
    staleTime: 30_000,
  });
};

export const useMember = (id: string | undefined) => {
  return useQuery<TeamMember | undefined>({
    queryKey: ['team', id],
    queryFn: async () => {
      if (!id) return undefined;
      const res = await unwrap(safe.get(`/team/members/${id}`));
      const m = res?.user ?? res;
      if (!m) return undefined;
      const parsed = teamMemberSchema.parse(m);
      return {
        ...parsed,
        name: parsed.fullName || parsed.name || parsed.email,
        apps: parsed.apps ?? [],
        status: parsed.isActive ? 'active' : 'inactive',
      } as TeamMember;
    },
    enabled: !!id,
  });
};

export const useInviteMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      email: string;
      fullName: string;
      role: TeamMember['role'];
      apps: string[];
      groupId?: number;
    }) => {
      const res = await unwrap(safe.post('/team/invite', payload));
      return res;
    },
    onSuccess: () => {
      toast.success('Invitation sent');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to send invitation'),
  });
};

export const useUpdateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, patch }: { id: string; patch: Partial<TeamMember> }) => {
      return unwrap(
        safe.put('/team/users/permissions', {
          userId: Number(id),
          apps: patch.apps ?? [],
          role: patch.role,
        })
      );
    },
    onSuccess: () => {
      toast.success('Member updated');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to update member'),
  });
};

export const useDeactivateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(safe.post(`/team/members/${id}/deactivate`));
      return id;
    },
    onSuccess: () => {
      toast.success('Member deactivated');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to deactivate member'),
  });
};
