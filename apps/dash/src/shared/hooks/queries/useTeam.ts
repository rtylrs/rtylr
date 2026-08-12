import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { MembersServiceClient } from '@voxire/voxie/hub/team/members/v1/client';
import { PermissionsServiceClient } from '@voxire/voxie/hub/team/permissions/v1/client';
import { InvitationsServiceClient } from '@voxire/voxie/hub/team/invitations/v1/client';
import { AccessServiceClient } from '@voxire/voxie/hub/team/access/v1/client';
import type { AccessRequest as ApiAccessRequest } from '@voxire/voxie/hub/team/access/v1/types';
import type {
  StaffUserPayload,
  StaffUserWithGroups,
} from '@voxire/voxie/hub/team/members/v1/types';
import { z } from 'zod';
import { toast } from 'sonner';
import { teamMemberSchema } from '../../schemas/team';

export type { TeamMember } from '../../schemas/team';
type TeamMember = z.infer<typeof teamMemberSchema> & { status: string };

const membersClient = createOneKitClient(MembersServiceClient, 'dash');
const permissionsClient = createOneKitClient(PermissionsServiceClient, 'dash');
const invitationsClient = createOneKitClient(InvitationsServiceClient, 'dash');
const accessClient = createOneKitClient(AccessServiceClient, 'dash');

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

function fromApiMember(u: StaffUserPayload | StaffUserWithGroups): TeamMember {
  const email = u.email ?? '';
  return {
    id: u.id ?? '',
    fullName: u.fullName,
    name: u.fullName || email,
    email,
    role: u.role ?? '',
    apps: u.apps ?? [],
    isActive: u.isActive ?? false,
    status: u.isActive ? 'active' : 'inactive',
    createdAt: u.createdAt ?? '',
    groups:
      'groups' in u
        ? (u.groups ?? []).map((g) => ({
            ...g,
            id: Number(g.id),
            name: g.name ?? '',
            color: g.color ?? '',
          }))
        : [],
  };
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

// hub.team.members.v1.ListTeam has no pagination support - the whole list is
// always returned in one call. page/pageSize are UI-only and echoed back.
export const useTeam = (params: UseTeamParams = {}) => {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<TeamListResult>({
    queryKey: ['team', { page, pageSize }],
    queryFn: async (): Promise<TeamListResult> => {
      const res = await membersClient.listTeam({});
      return {
        members: (res.users ?? []).map(fromApiMember),
        hasMore: false,
        page,
        pageSize,
      };
    },
  });
};

export const useMember = (id: string | undefined) => {
  return useQuery<TeamMember | undefined>({
    queryKey: ['team', id],
    queryFn: async () => {
      if (!id) return undefined;
      const res = await membersClient.getMember({ userId: id });
      return res.user ? fromApiMember(res.user) : undefined;
    },
    enabled: !!id,
  });
};

interface InviteMemberPayload {
  email: string;
  fullName: string;
  role: TeamMember['role'];
  apps: string[];
  groupId?: number;
  password?: string;
}

export const useInviteMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: InviteMemberPayload) => {
      return invitationsClient.inviteUser({
        email: payload.email,
        fullName: payload.fullName,
        role: payload.role,
        apps: payload.apps,
        ...(payload.groupId ? { groupId: String(payload.groupId) } : {}),
        ...(payload.password ? { password: payload.password } : {}),
      });
    },
    onSuccess: () => {
      toast.success('Invitation sent');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to send invitation')),
  });
};

export const useUpdateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, patch }: { id: string; patch: Partial<TeamMember> }) => {
      return permissionsClient.updatePermissions({
        userId: id,
        ...(patch.fullName || patch.name ? { fullName: patch.fullName ?? patch.name } : {}),
        apps: patch.apps ?? [],
        role: patch.role ?? 'staff',
      });
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
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to update member')),
  });
};

export const useDeactivateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await membersClient.removeMember({ userId: id });
      return id;
    },
    onSuccess: () => {
      toast.success('Member deactivated');
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to deactivate member')),
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

function fromApiAccessRequest(r: ApiAccessRequest): AccessRequest {
  return {
    id: Number(r.id),
    userId: Number(r.userId),
    app: r.app ?? '',
    requesterName: r.requesterName ?? '',
    requesterEmail: r.requesterEmail ?? '',
    createdAt: r.createdAt ?? '',
  };
}

export const useAccessRequests = () =>
  useQuery<AccessRequest[]>({
    queryKey: ['team', 'access-requests'],
    queryFn: async () => {
      const res = await accessClient.listAccessRequests({});
      return (res.requests ?? []).map(fromApiAccessRequest);
    },
    staleTime: 30_000,
  });

export const useApproveAccessRequest = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, userId, app }: { id: number; userId: number; app: string }) =>
      accessClient.approveAccessRequest({ id: String(id), userId: String(userId), app }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['team', 'access-requests'] });
      void qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Failed to grant access')),
  });
};
