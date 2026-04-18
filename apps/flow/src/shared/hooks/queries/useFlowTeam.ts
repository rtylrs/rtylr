import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../lib/api';

// ── Types ─────────────────────────────────────────────────────────────────────

export interface FlowTeamMember {
  id: string;
  name: string;
  email: string;
  role: string;
  apps: string[];
  isActive: boolean;
  avatarUrl?: string;
}

interface ListMembersResponse {
  users?: FlowTeamMember[];
  members?: FlowTeamMember[];
}

interface InvitePayload {
  email: string;
  fullName: string;
  role: string;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

export const useFlowTeam = (): ReturnType<typeof useQuery<FlowTeamMember[]>> =>
  useQuery<FlowTeamMember[]>({
    queryKey: ['flow-team'],
    queryFn: async (): Promise<FlowTeamMember[]> => {
      const res = await unwrap(safe.get<ListMembersResponse>('/team/members'));
      const raw: FlowTeamMember[] = res?.users ?? res?.members ?? [];
      return raw.map((m) => ({
        ...m,
        name: (m.name as string) || (m.email as string),
      }));
    },
    staleTime: 30_000,
  });

export const useFlowInviteMember = (): UseMutationResult<unknown, unknown, InvitePayload> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, InvitePayload>({
    mutationFn: (payload: InvitePayload): Promise<unknown> =>
      unwrap(safe.post('/auth/invite', { ...payload, apps: ['flow'] })),
    onSuccess: (): void => {
      toast.success('Invitation sent');
      void qc.invalidateQueries({ queryKey: ['flow-team'] });
    },
    onError: (): void => {
      toast.error('Failed to send invitation');
    },
  });
};

export const useFlowDeactivateMember = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> =>
      unwrap(safe.post(`/team/members/${id}/deactivate`, {})),
    onSuccess: (): void => {
      toast.success('Member removed from Flow');
      void qc.invalidateQueries({ queryKey: ['flow-team'] });
    },
    onError: (): void => {
      toast.error('Failed to remove member');
    },
  });
};
