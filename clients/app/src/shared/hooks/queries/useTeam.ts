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
import i18next from 'i18next';
import { parseEffectiveAccess } from '@voxire/auth';
import type { EffectiveAccess } from '@voxire/types';
import {
  permissionProfileAssignmentSchema,
  teamMemberSchema,
  type PermissionProfileAssignment,
} from '../../schemas/team';

export type TeamMember = z.infer<typeof teamMemberSchema> & {
  status: string;
  effectiveAccess?: EffectiveAccess;
  groupIds?: number[];
  isOwner?: boolean;
};

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
  total: number;
  page: number;
  pageSize: number;
}

function fromApiMember(u: StaffUserPayload | StaffUserWithGroups): TeamMember {
  const email = u.email ?? '';
  const effectiveAccess = parseEffectiveAccess(u.effectiveAccess);
  return {
    id: u.id ?? '',
    fullName: u.fullName,
    name: u.fullName || email,
    email,
    memberType: u.memberType ?? 'member',
    licensedAppIds: u.licensedAppIds ?? [],
    permissionProfileAssignments: z
      .array(permissionProfileAssignmentSchema)
      .parse(u.permissionProfileAssignments ?? []),
    isActive: u.isActive ?? false,
    status: u.isActive ? 'active' : 'inactive',
    effectiveAccess: effectiveAccess ?? undefined,
    isOwner: effectiveAccess?.isOwner ?? false,
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

export const useTeam = (params: UseTeamParams = {}) => {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<TeamListResult>({
    queryKey: ['team', { page, pageSize }],
    queryFn: async (): Promise<TeamListResult> => {
      const res = await membersClient.listTeam({ page, pageSize });
      const allMembers = (res.users ?? []).map(fromApiMember);
      const total =
        typeof res.total === 'number' && Number.isSafeInteger(res.total) && res.total >= 0
          ? res.total
          : allMembers.length;
      const hasMore = typeof res.total === 'number' ? page * pageSize < total : false;
      return {
        members: allMembers,
        hasMore,
        total,
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
  memberType: 'member' | 'admin';
  licensedAppIds: string[];
  permissionProfileAssignments: PermissionProfileAssignment[];
  username?: string;
  temporaryPassword?: string;
  groupIds?: number[];
  provisioningMode?: 'invitation' | 'direct_credentials';
}

export const useInviteMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: InviteMemberPayload) => {
      return invitationsClient.inviteUser({
        email: payload.email,
        fullName: payload.fullName,
        memberType: payload.memberType,
        licensedAppIds: payload.licensedAppIds,
        permissionProfileAssignments: payload.permissionProfileAssignments,
        ...(payload.groupIds?.length
          ? { groupIds: [...new Set(payload.groupIds.map(String))] }
          : {}),
        ...(payload.provisioningMode ? { provisioningMode: payload.provisioningMode } : {}),
        ...(payload.username ? { username: payload.username } : {}),
        ...(payload.temporaryPassword ? { temporaryPassword: payload.temporaryPassword } : {}),
      });
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.invitationSent', 'Invitation sent'));
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.failedSendInvitation', 'Failed to send invitation')
        )
      ),
  });
};

export const useUpdateMember = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, patch }: { id: string; patch: Partial<TeamMember> }) => {
      return permissionsClient.updatePermissions({
        userId: id,
        ...(patch.fullName || patch.name ? { fullName: patch.fullName ?? patch.name } : {}),
        ...(patch.memberType ? { memberType: patch.memberType } : {}),
        replaceLicensedAppIds: patch.licensedAppIds !== undefined,
        ...(patch.licensedAppIds !== undefined ? { licensedAppIds: patch.licensedAppIds } : {}),
        replacePermissionProfileAssignments: patch.permissionProfileAssignments !== undefined,
        ...(patch.permissionProfileAssignments !== undefined
          ? { permissionProfileAssignments: patch.permissionProfileAssignments }
          : {}),
        replaceGroupIds: patch.groupIds !== undefined,
        ...(patch.groupIds ? { groupIds: patch.groupIds.map(String) } : {}),
      });
    },
    onSuccess: (_data, { id, patch }) => {
      toast.success(i18next.t('common.toasts.memberUpdated', 'Member updated'));
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
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.failedUpdateMember', 'Failed to update member')
        )
      ),
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
      toast.success(i18next.t('common.toasts.memberDeactivated', 'Member deactivated'));
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.failedDeactivateMember', 'Failed to deactivate member')
        )
      ),
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
      const res = (await accessClient.listAccessRequests({})) as {
        requests?: ApiAccessRequest[];
      };
      return (res.requests ?? []).map((r) => ({
        id: Number(r.id),
        userId: Number(r.userId),
        app: r.app ?? '',
        requesterName: r.requesterName ?? '',
        requesterEmail: r.requesterEmail ?? '',
        createdAt: r.createdAt ?? '',
      }));
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
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.failedGrantAccess', 'Failed to grant access')
        )
      ),
  });
};
