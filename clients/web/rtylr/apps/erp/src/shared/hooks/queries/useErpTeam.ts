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

export async function fetchAllErpTeam(pageSize = 100): Promise<ErpTeamMember[]> {
  const members: ErpTeamMember[] = [];
  const seenPages = new Set<string>();
  let page = 1;

  while (true) {
    const res = await unwrap(oneKitResult(() => teamMembersClient.listTeam({ page, pageSize })));
    const raw = res.users ?? [];
    const pageMembers = raw.map((member) => ({
      id: member.id ?? '',
      name: member.fullName || member.email || '',
      email: member.email ?? '',
      role: member.role ?? '',
      apps: member.apps ?? [],
      isActive: member.isActive ?? false,
    }));
    const signature = pageMembers.map((member) => member.id || member.email).join('|');
    if (pageMembers.length > 0 && seenPages.has(signature)) {
      throw new Error('Team pagination did not advance');
    }
    if (pageMembers.length > 0) seenPages.add(signature);
    members.push(...pageMembers);

    const hasMore =
      res.total !== undefined ? members.length < res.total : pageMembers.length === pageSize;
    if (!hasMore || pageMembers.length === 0) return members;
    page += 1;
  }
}

export const useErpTeam = (): UseQueryResult<ErpTeamMember[]> =>
  useQuery<ErpTeamMember[]>({
    queryKey: ['erp-team'],
    queryFn: () => fetchAllErpTeam(),
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
