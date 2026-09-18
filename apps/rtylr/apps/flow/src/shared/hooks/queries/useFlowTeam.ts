import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
} from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { MembersServiceClient } from '@voxire/voxie/hub/team/members/v1/client';
import { PermissionsServiceClient } from '@voxire/voxie/hub/team/permissions/v1/client';
import type { StaffUserPayload } from '@voxire/voxie/hub/team/members/v1/types';

const membersClient = createOneKitClient(MembersServiceClient, 'flow');
const permissionsClient = createOneKitClient(PermissionsServiceClient, 'flow');

// ── Types ─────────────────────────────────────────────────────────────────────

export interface FlowTeamMember {
  id: string;
  name: string;
  fullName?: string;
  email: string;
  role: string;
  apps: string[];
  isActive: boolean;
  avatarUrl?: string;
}

interface GrantPayload {
  userId: string;
  // Member's existing apps and role come from useFlowTeam so the
  // permissions update is additive - we only flip Flow on, nothing
  // else is touched.
  apps: string[];
  role: string;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

export function normalizeFlowTeamMember(member: StaffUserPayload): FlowTeamMember {
  const fullName = member.fullName?.trim() ?? '';
  const email = member.email?.trim() ?? '';

  return {
    id: member.id ?? '',
    name: fullName || email,
    fullName: fullName || undefined,
    email,
    role: member.role ?? '',
    apps: member.apps ?? [],
    isActive: member.isActive ?? false,
  };
}

// useFlowTeam returns every member of the business. Flow-specific
// filters (current members vs. grant candidates) are applied in the UI
// against each member's `.apps` array - there is no separate Flow-only
// endpoint, so the same query backs both lists.
export const useFlowTeam = (): ReturnType<typeof useQuery<FlowTeamMember[]>> =>
  useQuery<FlowTeamMember[]>({
    queryKey: ['flow-team'],
    queryFn: async (): Promise<FlowTeamMember[]> => {
      const res = await membersClient.listTeam({});
      return (res.users ?? []).map(normalizeFlowTeamMember);
    },
    staleTime: 30_000,
  });

// useFlowGrantAccess grants Flow access to an existing business member.
// External invites happen at the business level (hub Team page); here
// we only flip the bit for someone already in the org.
export const useFlowGrantAccess = (): UseMutationResult<unknown, unknown, GrantPayload> => {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<unknown, unknown, GrantPayload>({
    mutationFn: ({ userId, apps, role }: GrantPayload): Promise<unknown> => {
      const nextApps = apps.includes('flow') ? apps : [...apps, 'flow'];
      return permissionsClient.updatePermissions({ userId, apps: nextApps, role });
    },
    onSuccess: (): void => {
      toast.success(t('toasts.flowTeam.accessGranted'));
      void qc.invalidateQueries({ queryKey: ['flow-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('toasts.flowTeam.grantAccessFailed')));
    },
  });
};

export const useFlowDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> => membersClient.removeMember({ userId: id }),
    onSuccess: (): void => {
      toast.success(t('toasts.flowTeam.memberRemoved'));
      void qc.invalidateQueries({ queryKey: ['flow-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('toasts.flowTeam.removeMemberFailed')));
    },
  });
};
