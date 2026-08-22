import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { MembersServiceClient } from '@voxire/voxie/hub/team/members/v1/client';
import type { StaffUserPayload } from '@voxire/voxie/hub/team/members/v1/types';
import { InvitationsServiceClient } from '@voxire/voxie/hub/team/invitations/v1/client';

export type CrmTeamMember = StaffUserPayload;

const membersClient = createOneKitClient(MembersServiceClient, 'crm');
const invitationsClient = createOneKitClient(InvitationsServiceClient, 'crm');

interface InvitePayload {
  email: string;
  fullName: string;
  role: string;
  password?: string;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

export const useCrmTeam = (): UseQueryResult<CrmTeamMember[]> =>
  useQuery<CrmTeamMember[]>({
    queryKey: ['crm-team'],
    queryFn: async (): Promise<CrmTeamMember[]> => {
      const res = await membersClient.listTeam({});
      return res.users ?? [];
    },
    staleTime: 30_000,
  });

export const useCrmInviteMember = (): UseMutationResult<unknown, unknown, InvitePayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, InvitePayload>({
    mutationFn: (payload: InvitePayload) =>
      invitationsClient.inviteUser({
        ...payload,
        apps: ['crm'],
        ...(payload.password ? { password: payload.password } : {}),
      }),
    onSuccess: (): void => {
      toast.success(t('crm.team.inviteSent'));
      void qc.invalidateQueries({ queryKey: ['crm-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('crm.team.inviteFailed')));
    },
  });
};

export const useCrmDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string) => membersClient.removeMember({ userId: id }),
    onSuccess: (): void => {
      toast.success(t('crm.team.memberRemoved'));
      void qc.invalidateQueries({ queryKey: ['crm-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('crm.team.removeFailed')));
    },
  });
};
