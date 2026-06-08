import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { safe } from '../../lib/api';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface CrmTeamMember {
  id: string;
  name: string;
  email: string;
  role: string;
  apps: string[];
  isActive: boolean;
}

interface ListMembersResponse {
  users?: CrmTeamMember[];
  members?: CrmTeamMember[];
}

interface InvitePayload {
  email: string;
  fullName: string;
  role: string;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

export const useCrmTeam = (): UseQueryResult<CrmTeamMember[]> =>
  useQuery<CrmTeamMember[]>({
    queryKey: ['crm-team'],
    queryFn: async (): Promise<CrmTeamMember[]> => {
      const res = await unwrap(safe.get<ListMembersResponse>('/team/members'));
      const raw: CrmTeamMember[] = res?.users ?? res?.members ?? [];
      return raw.map((m) => ({ ...m, name: (m.name as string) || (m.email as string) }));
    },
    staleTime: 30_000,
  });

export const useCrmInviteMember = (): UseMutationResult<unknown, unknown, InvitePayload> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, InvitePayload>({
    mutationFn: (payload: InvitePayload): Promise<unknown> =>
      unwrap(safe.post('/team/invite', { ...payload, apps: ['crm'] })),
    onSuccess: (): void => {
      toast.success(t('crm.team.inviteSent'));
      void qc.invalidateQueries({ queryKey: ['crm-team'] });
    },
    onError: (): void => {
      toast.error(t('crm.team.inviteFailed'));
    },
  });
};

export const useCrmDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> =>
      unwrap(safe.patch(`/team/members/${id}/deactivate`, {})),
    onSuccess: (): void => {
      toast.success(t('crm.team.memberRemoved'));
      void qc.invalidateQueries({ queryKey: ['crm-team'] });
    },
    onError: (): void => {
      toast.error(t('crm.team.removeFailed'));
    },
  });
};
