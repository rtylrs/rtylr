import { unwrap, createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../lib/api';
import { InvitationsServiceClient } from '@voxire/voxie/hub/team/invitations/v1/client';

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

// Hooks
export const useGroups = (): ReturnType<typeof useQuery<Group[]>> =>
  useQuery<Group[]>({
    queryKey: ['groups'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/api/hub/team/groups/v1/groups'));
      return (res as { groups?: Group[] })?.groups ?? [];
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
      const res = await unwrap(safe.get(`/api/hub/team/groups/v1/groups/${id}`));
      return (res as { group?: GroupDetail })?.group;
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
      const res = await unwrap(safe.post('/api/hub/team/groups/v1/groups', payload));
      return (res as { group: Group })?.group;
    },
    onSuccess: () => {
      toast.success('Group created');
      qc.invalidateQueries({ queryKey: ['groups'] });
    },
    onError: () => toast.error('Failed to create group'),
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
      const res = await unwrap(safe.put(`/api/hub/team/groups/v1/groups/${id}`, payload));
      return (res as { group: Group })?.group;
    },
    onSuccess: () => {
      toast.success('Group saved');
      qc.invalidateQueries({ queryKey: ['groups'] });
      qc.invalidateQueries({ queryKey: ['groups', id] });
    },
    onError: () => toast.error('Failed to save group'),
  });
};

export const useDeleteGroup = (): ReturnType<typeof useMutation<void, Error, number>> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (groupId: number) => {
      await unwrap(safe.delete(`/api/hub/team/groups/v1/groups/${groupId}`));
    },
    onSuccess: () => {
      toast.success('Group deleted');
      qc.invalidateQueries({ queryKey: ['groups'] });
    },
    onError: () => toast.error('Failed to delete group'),
  });
};

export const useAddGroupMembers = (
  groupId: string | undefined
): ReturnType<typeof useMutation<void, Error, number[]>> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userIds: number[]) => {
      await unwrap(
        safe.post(`/api/hub/team/groups/v1/groups/${groupId}/members`, { user_ids: userIds })
      );
    },
    onSuccess: () => {
      toast.success('Members added');
      qc.invalidateQueries({ queryKey: ['groups'] });
      qc.invalidateQueries({ queryKey: ['groups', groupId] });
    },
    onError: () => toast.error('Failed to add members'),
  });
};

export const useRemoveGroupMember = (
  groupId: string | undefined
): ReturnType<typeof useMutation<void, Error, number>> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userId: number) => {
      await unwrap(safe.delete(`/api/hub/team/groups/v1/groups/${groupId}/members/${userId}`));
    },
    onSuccess: () => {
      toast.success('Member removed');
      qc.invalidateQueries({ queryKey: ['groups', groupId] });
      qc.invalidateQueries({ queryKey: ['team'] });
    },
    onError: () => toast.error('Failed to remove member'),
  });
};

export const useBulkInvite = (): ReturnType<
  typeof useMutation<
    { results: BulkResult[]; message: string },
    Error,
    { emails: string[]; role: string; apps: string[]; groupId?: number }
  >
> => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: {
      emails: string[];
      role: string;
      apps: string[];
      groupId?: number;
    }) => {
      const res = await invitationsClient.bulkInvite({
        emails: payload.emails,
        role: payload.role,
        apps: payload.apps,
        groupId: payload.groupId ? String(payload.groupId) : '',
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
    onError: (error) => toast.error(getOneKitErrorMessage(error, 'Bulk invite failed')),
  });
};
