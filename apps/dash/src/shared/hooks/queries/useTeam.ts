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

function applyMemberPatch(member: TeamMember, patch: Partial<TeamMember>): TeamMember {
  const fullName = patch.fullName ?? patch.name;
  const next: TeamMember = {
    ...member,
    ...patch,
  };

  if (fullName !== undefined) {
    next.fullName = fullName;
    next.name = fullName || next.email;
  } else {
    next.name = next.fullName || next.name || next.email;
  }

  if (patch.isActive !== undefined) {
    next.status = patch.isActive ? 'active' : 'inactive';
  }

  return next;
}

function isTeamListResult(value: unknown): value is TeamListResult {
  return (
    typeof value === 'object' && value !== null && Array.isArray((value as TeamListResult).members)
  );
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
      password?: string;
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
      const payload: Record<string, unknown> = { userId: Number(id) };
      if (patch.fullName !== undefined || patch.name !== undefined) {
        payload.fullName = patch.fullName ?? patch.name ?? '';
      }
      if (patch.apps !== undefined) payload.apps = patch.apps;
      if (patch.role !== undefined) payload.role = patch.role;

      return unwrap(safe.put('/team/users/permissions', payload));
    },
    onSuccess: (_data, { id, patch }) => {
      toast.success('Member updated');
      qc.getQueriesData({ queryKey: ['team'] }).forEach(([queryKey, cached]) => {
        if (!isTeamListResult(cached)) return;
        qc.setQueryData<TeamListResult>(queryKey, {
          ...cached,
          members: cached.members.map((member) =>
            member.id === id ? applyMemberPatch(member, patch) : member
          ),
        });
      });
      qc.setQueryData<TeamMember | undefined>(['team', id], (old) =>
        old ? applyMemberPatch(old, patch) : old
      );
    },
    onError: () => toast.error('Failed to update member'),
  });
};

export const useDeactivateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(safe.patch(`/team/members/${id}/deactivate`));
      return id;
    },
    onSuccess: () => {
      toast.success('Member deactivated');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to deactivate member'),
  });
};

export interface AccessRequest {
  id: number;
  userId: number;
  app: string;
  requesterName: string;
  requesterEmail: string;
  createdAt: string;
}

export const useAccessRequests = () =>
  useQuery<AccessRequest[]>({
    queryKey: ['team', 'access-requests'],
    queryFn: async () => {
      const res = await unwrap(safe.get<{ requests: AccessRequest[] }>('/team/access-requests'));
      return res?.requests ?? [];
    },
    staleTime: 30_000,
  });

export const useApproveAccessRequest = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, userId, app }: { id: number; userId: number; app: string }) =>
      unwrap(safe.post(`/team/access-requests/${id}/approve`, { userId, app })),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['team', 'access-requests'] });
      void qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to grant access'),
  });
};
