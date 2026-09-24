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

const CRM_TEAM_PAGE_SIZE = 50;

async function fetchAllCrmTeamMembers(): Promise<CrmTeamMember[]> {
  const all: CrmTeamMember[] = [];
  const seenPages = new Set<string>();
  let page = 1;

  for (;;) {
    const res = await membersClient.listTeam({ page, pageSize: CRM_TEAM_PAGE_SIZE });
    const users = res.users ?? [];
    if (users.length === 0) {
      if (typeof res.total === 'number' && all.length < res.total) {
        throw new Error('CRM team pagination ended before the reported total');
      }
      return all;
    }

    const pageKey = users
      .map((user) => user.id ?? `${user.email ?? ''}:${user.fullName ?? ''}`)
      .join('|');
    if (seenPages.has(pageKey)) {
      throw new Error('CRM team pagination did not advance');
    }
    seenPages.add(pageKey);
    all.push(...users);

    const hasNext =
      typeof res.total === 'number' ? all.length < res.total : users.length === CRM_TEAM_PAGE_SIZE;
    if (!hasNext) return all;

    page += 1;
  }
}

export const useCrmTeam = (): UseQueryResult<CrmTeamMember[]> =>
  useQuery<CrmTeamMember[]>({
    queryKey: ['crm-team'],
    queryFn: fetchAllCrmTeamMembers,
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
