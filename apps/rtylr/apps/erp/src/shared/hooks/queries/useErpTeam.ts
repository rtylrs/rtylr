import { createOneKitClient, getOneKitErrorMessage, oneKitResult, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseMutationResult, UseQueryResult } from '@tanstack/react-query';
import { toast } from 'sonner';
import { InvitationsServiceClient } from '@voxire/voxie/hub/team/invitations/v1/client';
import { MembersServiceClient } from '@voxire/voxie/hub/team/members/v1/client';

const teamMembersClient = createOneKitClient(MembersServiceClient, 'erp');
const teamInvitationsClient = createOneKitClient(InvitationsServiceClient, 'erp');

// ── Types ─────────────────────────────────────────────────────────────────────

export interface ErpTeamMember {
  id: string;
  name: string;
  email: string;
  role: string;
  apps: string[];
  isActive: boolean;
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
      const res = await unwrap(oneKitResult(() => teamMembersClient.listTeam({})));
      const raw = res.users ?? [];
      return raw.map((m) => ({
        id: m.id ?? '',
        name: m.fullName || m.email || '',
        email: m.email ?? '',
        role: m.role ?? '',
        apps: m.apps ?? [],
        isActive: m.isActive ?? false,
      }));
    },
    staleTime: 30_000,
  });

export const useErpInviteMember = (): UseMutationResult<unknown, unknown, InvitePayload> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, InvitePayload>({
    mutationFn: (payload: InvitePayload): Promise<unknown> =>
      unwrap(oneKitResult(() => teamInvitationsClient.inviteUser({ ...payload, apps: ['erp'] }))),
    onSuccess: (): void => {
      toast.success('Invitation sent');
      void qc.invalidateQueries({ queryKey: ['erp-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, 'Failed to send invitation'));
    },
  });
};

export const useErpDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> =>
      unwrap(oneKitResult(() => teamMembersClient.removeMember({ userId: id }))),
    onSuccess: (): void => {
      toast.success('Member removed');
      void qc.invalidateQueries({ queryKey: ['erp-team'] });
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, 'Failed to remove member'));
    },
  });
};
