import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { rest } from '../../lib/clients';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface ErpTeamMember {
  id: string;
  name: string;
  email: string;
  role: string;
  apps: string[];
  isActive: boolean;
}

interface ListMembersResponse {
  users?: ErpTeamMember[];
  members?: ErpTeamMember[];
}

interface InvitePayload {
  email: string;
  fullName: string;
  role: string;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

export const useErpTeam = (): UseQueryResult<ErpTeamMember[]> =>
  useQuery<ErpTeamMember[]>({
    queryKey: ['erp-team'],
    queryFn: async (): Promise<ErpTeamMember[]> => {
      const res = await unwrap(rest.get<ListMembersResponse>('/team/members'));
      const raw: ErpTeamMember[] = res?.users ?? res?.members ?? [];
      return raw.map((m) => ({ ...m, name: (m.name as string) || (m.email as string) }));
    },
    staleTime: 30_000,
  });

export const useErpInviteMember = (): UseMutationResult<unknown, unknown, InvitePayload> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, InvitePayload>({
    mutationFn: (payload: InvitePayload): Promise<unknown> =>
      unwrap(rest.post('/team/invite', { ...payload, apps: ['erp'] })),
    onSuccess: (): void => {
      toast.success('Invitation sent');
      void qc.invalidateQueries({ queryKey: ['erp-team'] });
    },
    onError: (): void => {
      toast.error('Failed to send invitation');
    },
  });
};

export const useErpDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> =>
      unwrap(rest.patch(`/team/members/${id}/deactivate`, {})),
    onSuccess: (): void => {
      toast.success('Member removed');
      void qc.invalidateQueries({ queryKey: ['erp-team'] });
    },
    onError: (): void => {
      toast.error('Failed to remove member');
    },
  });
};
