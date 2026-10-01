import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { MembersServiceClient } from '@voxire/voxie/hub/team/members/v1/client';
import type { StaffUserPayload } from '@voxire/voxie/hub/team/members/v1/types';
import { hasEffectiveAppAccess, parseEffectiveAccess } from '@voxire/auth';

const membersClient = createOneKitClient(MembersServiceClient, 'flow');

export interface FlowTeamMember {
  id: string;
  name: string;
  fullName?: string;
  email: string;
  memberType: string;
  isOwner: boolean;
  hasFlowAccess: boolean;
  isActive: boolean;
  avatarUrl?: string;
}

export function normalizeFlowTeamMember(member: StaffUserPayload): FlowTeamMember {
  const fullName = member.fullName?.trim() ?? '';
  const email = member.email?.trim() ?? '';
  const effectiveAccess = parseEffectiveAccess(member.effectiveAccess);

  return {
    id: String(member.id ?? ''),
    name: fullName || email,
    fullName: fullName || undefined,
    email,
    memberType: member.memberType ?? 'member',
    isOwner: effectiveAccess?.isOwner ?? false,
    hasFlowAccess: hasEffectiveAppAccess(effectiveAccess, 'flow'),
    isActive: member.isActive ?? false,
  };
}

/** Flow consumes the shared business team and filters it by resolved access. */
export const useFlowTeam = (): ReturnType<typeof useQuery<FlowTeamMember[]>> =>
  useQuery<FlowTeamMember[]>({
    queryKey: ['flow-team'],
    queryFn: async (): Promise<FlowTeamMember[]> => {
      const res = await membersClient.listTeam({});
      return (res.users ?? []).map(normalizeFlowTeamMember);
    },
    staleTime: 30_000,
  });
