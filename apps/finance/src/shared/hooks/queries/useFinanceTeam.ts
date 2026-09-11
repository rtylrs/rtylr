import { getOneKitErrorMessage, oneKitResult, unwrap } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { teamInvitationsClient, teamMembersClient } from '../../lib/clients';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface FinanceTeamMember {
  id: string;
  name: string;
  email: string;
  role: string;
  apps: string[];
  isActive: boolean;
}

interface ListMembersResponse {
  users?: FinanceTeamMember[];
  members?: FinanceTeamMember[];
}

interface InvitePayload {
  email: string;
  fullName: string;
  role: string;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

export const useFinanceTeam = (): UseQueryResult<FinanceTeamMember[]> =>
  useQuery<FinanceTeamMember[]>({
    queryKey: ['finance-team'],
    queryFn: async (): Promise<FinanceTeamMember[]> => {
      const res = (await unwrap(
        oneKitResult(() => teamMembersClient.listTeam({}))
      )) as ListMembersResponse;
      const raw: FinanceTeamMember[] = res?.users ?? res?.members ?? [];
      return raw.map((m) => ({ ...m, name: (m.name as string) || (m.email as string) }));
    },
    staleTime: 30_000,
  });

export const useFinanceInviteMember = (): UseMutationResult<unknown, unknown, InvitePayload> => {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<unknown, unknown, InvitePayload>({
    mutationFn: (payload: InvitePayload): Promise<unknown> =>
      unwrap(
        oneKitResult(() => teamInvitationsClient.inviteUser({ ...payload, apps: ['finance'] }))
      ),
    onSuccess: (): void => {
      toast.success(t('finance.team.invited'));
      void qc.invalidateQueries({ queryKey: ['finance-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, 'Failed to send invitation'));
    },
  });
};

export const useFinanceDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> =>
      // removeMember, not a deactivate call. No route named deactivate exists
      // in onion or in any contract, so the previous
      // PATCH /team/members/{id}/deactivate had always 404'd (#910).
      unwrap(oneKitResult(() => teamMembersClient.removeMember({ userId: id }))),
    onSuccess: (): void => {
      toast.success(t('finance.team.removed'));
      void qc.invalidateQueries({ queryKey: ['finance-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, 'Failed to remove member'));
    },
  });
};
