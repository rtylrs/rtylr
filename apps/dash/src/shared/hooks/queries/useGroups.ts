import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import i18next from 'i18next';
import { groupsClient } from '../../lib/clients';
import { InvitationsServiceClient } from '@voxire/voxie/hub/team/invitations/v1/client';
import type {
  Group as ApiGroup,
  GroupDetail as ApiGroupDetail,
} from '@voxire/voxie/hub/team/groups/v1/types';
import type { AppRoleAssignment } from '@voxire/voxie/hub/user/v1/types';

const invitationsClient = createOneKitClient(InvitationsServiceClient, 'dash');

// Types
export type ResourcePermissionLevel = 'viewer' | 'editor';

export type ResourcePermission = {
  resourceKey: string;
  permissionLevel: ResourcePermissionLevel;
};

export type GroupSummary = {
  id: number;
  name: string;
  color: string;
};

export type Group = {
  id: number;
  ownerId: number;
  name: string;
  description: string;
  color: string;
  resourcePermissions: ResourcePermission[];
  memberCount: number;
  createdAt: string;
  updatedAt: string;
};

export type GroupDetail = Group & {
  members: GroupMember[];
};

export type GroupMember = {
  userId: number;
  email: string;
  fullName: string;
  addedAt: string;
};

export type BulkResult = {
  email: string;
  userId: number;
  status: 'invited' | 'existing' | 'error';
  message: string;
};

// The wire carries int64 ids as strings; normalize once here so every
// consumer can treat ids/userIds as numbers without re-parsing.
function toIdNumber(value: string | undefined): number {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}

function toResourcePermissions(perms: ApiGroup['resourcePermissions']): ResourcePermission[] {
  return (perms ?? []).map((p) => ({
    resourceKey: p.resourceKey ?? '',
    permissionLevel: p.permissionLevel === 'editor' ? 'editor' : 'viewer',
  }));
}

function fromApiGroup(g: ApiGroup): Group {
  return {
    id: toIdNumber(g.id),
    ownerId: toIdNumber(g.ownerId),
    name: g.name ?? '',
    description: g.description ?? '',
    color: g.color ?? '',
    resourcePermissions: toResourcePermissions(g.resourcePermissions),
    memberCount: g.memberCount ?? 0,
    createdAt: g.createdAt ?? '',
    updatedAt: g.updatedAt ?? '',
  };
}

function fromApiGroupDetail(g: ApiGroupDetail): GroupDetail {
  return {
    ...fromApiGroup(g),
    members: (g.members ?? []).map((m) => ({
      userId: toIdNumber(m.userId),
      email: m.email ?? '',
      fullName: m.fullName ?? '',
      addedAt: m.addedAt ?? '',
    })),
  };
}

// Hooks
export const useGroups = (): ReturnType<typeof useQuery<Group[]>> =>
  useQuery<Group[]>({
    queryKey: ['groups'],
    queryFn: async () => {
      const res = await groupsClient.listGroups({});
      return (res.groups ?? []).map(fromApiGroup);
    },
    staleTime: 30_000,
  });

export const useGroup = (
  id: string | undefined
): ReturnType<typeof useQuery<GroupDetail | undefined>> =>
  useQuery<GroupDetail | undefined>({
    queryKey: ['groups', id],
    queryFn: async () => {
      if (!id) return undefined;
      const res = await groupsClient.getGroupDetail({ id });
      return res.group ? fromApiGroupDetail(res.group) : undefined;
    },
    enabled: !!id,
  });

export const useCreateGroup = (): ReturnType<
  typeof useMutation<
    Group,
    Error,
    {
      name: string;
      description?: string;
      color?: string;
      resourcePermissions: ResourcePermission[];
    }
  >
> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      name: string;
      description?: string;
      color?: string;
      resourcePermissions: ResourcePermission[];
    }) => {
      const res = await groupsClient.createGroup(payload);
      if (!res.group) throw new Error('Group was not returned by the server');
      return fromApiGroup(res.group);
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.groupCreated', 'Group created'));
      qc.invalidateQueries({ queryKey: ['groups'] });
    },
    onError: () =>
      toast.error(i18next.t('common.toasts.failedCreateGroup', 'Failed to create group')),
  });
};

export const useUpdateGroup = (
  id: string | undefined
): ReturnType<
  typeof useMutation<
    Group,
    Error,
    {
      name: string;
      description?: string;
      color?: string;
      resourcePermissions: ResourcePermission[];
    }
  >
> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      name: string;
      description?: string;
      color?: string;
      resourcePermissions: ResourcePermission[];
    }) => {
      if (!id) throw new Error('Group id is required');
      const res = await groupsClient.updateGroup({ id, ...payload });
      if (!res.group) throw new Error('Group was not returned by the server');
      return fromApiGroup(res.group);
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.groupSaved', 'Group saved'));
      qc.invalidateQueries({ queryKey: ['groups'] });
      qc.invalidateQueries({ queryKey: ['groups', id] });
    },
    onError: () => toast.error(i18next.t('common.toasts.failedSaveGroup', 'Failed to save group')),
  });
};

export const useDeleteGroup = (): ReturnType<typeof useMutation<void, Error, number>> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (groupId: number) => {
      await groupsClient.deleteGroup({ id: String(groupId) });
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.groupDeleted', 'Group deleted'));
      qc.invalidateQueries({ queryKey: ['groups'] });
    },
    onError: () =>
      toast.error(i18next.t('common.toasts.failedDeleteGroup', 'Failed to delete group')),
  });
};

export const useAddGroupMembers = (
  groupId: string | undefined
): ReturnType<typeof useMutation<void, Error, (string | number)[]>> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userIds: (string | number)[]) => {
      if (!groupId) throw new Error('Group id is required');
      await groupsClient.addGroupMembers({
        id: groupId,
        // The wire carries int64 user ids as strings.
        userIds: userIds.map(String),
      });
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.membersAdded', 'Members added'));
      qc.invalidateQueries({ queryKey: ['groups'] });
      qc.invalidateQueries({ queryKey: ['groups', groupId] });
      // Team rows embed each member's group memberships; refresh them too.
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () =>
      toast.error(i18next.t('common.toasts.failedAddMembers', 'Failed to add members')),
  });
};

export const useRemoveGroupMember = (
  groupId: string | undefined
): ReturnType<typeof useMutation<void, Error, number>> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userId: number) => {
      if (!groupId) throw new Error('Group id is required');
      await groupsClient.removeGroupMember({ id: groupId, userId: String(userId) });
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.memberRemoved', 'Member removed'));
      qc.invalidateQueries({ queryKey: ['groups', groupId] });
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () =>
      toast.error(i18next.t('common.toasts.failedRemoveMember', 'Failed to remove member')),
  });
};

export const useBulkInvite = (): ReturnType<
  typeof useMutation<
    { results: BulkResult[]; message: string },
    Error,
    {
      emails: string[];
      role: string;
      apps: string[];
      groupId?: number;
      dashRole?: string;
      appAssignments?: AppRoleAssignment[];
      groupIds?: number[];
    }
  >
> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      emails: string[];
      role: string;
      apps: string[];
      groupId?: number;
      dashRole?: string;
      appAssignments?: AppRoleAssignment[];
      groupIds?: number[];
    }) => {
      const res = await invitationsClient.bulkInvite({
        emails: payload.emails,
        role: payload.role,
        apps: payload.apps,
        ...(payload.dashRole ? { dashRole: payload.dashRole } : {}),
        ...(payload.appAssignments ? { appAssignments: payload.appAssignments } : {}),
        ...(payload.groupIds ? { groupIds: payload.groupIds.map(String) } : {}),
        ...(payload.groupId ? { groupId: String(payload.groupId) } : {}),
      });
      return {
        results: (res.results ?? []).map((r) => ({
          email: r.email ?? '',
          userId: Number(r.userId),
          status: r.status as BulkResult['status'],
          message: r.message ?? '',
        })),
        message: res.message ?? '',
      };
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['team'] });
      qc.invalidateQueries({ queryKey: ['groups'] });
    },
    onError: (error) =>
      toast.error(
        getOneKitErrorMessage(
          error,
          i18next.t('common.toasts.bulkInviteFailed', 'Bulk invite failed')
        )
      ),
  });
};
