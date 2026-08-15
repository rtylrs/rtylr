import { getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../lib/api';

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
      const res = await unwrap(safe.get<ListMembersResponse>('/team/members'));
      const raw: FinanceTeamMember[] = res?.users ?? res?.members ?? [];
      return raw.map((m) => ({ ...m, name: (m.name as string) || (m.email as string) }));
    },
    staleTime: 30_000,
  });

export const useFinanceInviteMember = (): UseMutationResult<unknown, unknown, InvitePayload> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, InvitePayload>({
    mutationFn: (payload: InvitePayload): Promise<unknown> =>
      unwrap(safe.post('/team/invite', { ...payload, apps: ['finance'] })),
    onSuccess: (): void => {
      toast.success('Invitation sent');
      void qc.invalidateQueries({ queryKey: ['finance-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, 'Failed to send invitation'));
    },
  });
};

export const useFinanceDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> =>
      unwrap(safe.patch(`/team/members/${id}/deactivate`, {})),
    onSuccess: (): void => {
      toast.success('Member removed');
      void qc.invalidateQueries({ queryKey: ['finance-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, 'Failed to remove member'));
    },
  });
};
