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

interface GrantPayload {
  userId: string;
  // Member's existing apps and role come from useFlowTeam so the
  // permissions update is additive - we only flip Flow on, nothing
  // else is touched.
  apps: string[];
  role: string;
}

// ── Hooks ─────────────────────────────────────────────────────────────────────

// useFlowTeam returns every member of the business. Flow-specific
// filters (current members vs. grant candidates) are applied in the UI
// against each member's `.apps` array - there is no separate Flow-only
// endpoint, so the same query backs both lists.
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

// useFlowGrantAccess grants Flow access to an existing business member.
// External invites happen at the business level (hub Team page); here
// we only flip the bit for someone already in the org.
export const useFlowGrantAccess = (): UseMutationResult<unknown, unknown, GrantPayload> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, GrantPayload>({
    mutationFn: ({ userId, apps, role }: GrantPayload): Promise<unknown> => {
      const nextApps = apps.includes('flow') ? apps : [...apps, 'flow'];
      return unwrap(
        safe.put('/team/users/permissions', {
          user_id: Number(userId),
          apps: nextApps,
          role,
        })
      );
    },
    onSuccess: (): void => {
      toast.success('Flow access granted');
      void qc.invalidateQueries({ queryKey: ['flow-team'] });
    },
    onError: (): void => {
      toast.error('Failed to grant Flow access');
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
