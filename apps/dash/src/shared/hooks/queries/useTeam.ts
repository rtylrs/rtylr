import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { StaffUserPayload } from '@voxire/contracts/hub/v1/team';
import { z } from 'zod';
import api, { safe } from '../../lib/api';
import { teamMemberSchema } from '../../schemas/team';
import { toast } from 'sonner';

export type { TeamMember } from '../../schemas/team';
type TeamMember = z.infer<typeof teamMemberSchema> & { status: string };

export const useTeam = () => {
  return useQuery<TeamMember[]>({
    queryKey: ['team'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/team/members'));
      const raw = res?.users ?? res?.members ?? res;
      if (!Array.isArray(raw)) return [];
      return raw.map((m: StaffUserPayload) => {
        const parsed = teamMemberSchema.parse(m);
        return {
          ...parsed,
          name: parsed.fullName || parsed.name || parsed.email,
          status: parsed.isActive ? 'active' : 'inactive',
          groups: parsed.groups ?? [],
        } as TeamMember;
      });
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
      const res = await api.patch(`/team/members/${id}`, patch);
      return res;
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
