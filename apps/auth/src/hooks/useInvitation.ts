import { useMutation, useQuery } from '@tanstack/react-query';

export interface InvitationInfo {
  token: string;
  email: string;
  organization: string;
  role: string;
  invitedBy: string;
  expiresAt: string;
  expired?: boolean;
}

export const useInvitation = (token?: string) =>
  useQuery<InvitationInfo>({
    queryKey: ['invitation', token] as const,
    enabled: !!token,
    queryFn: async () => {
      await new Promise((resolve) => setTimeout(resolve, 300));
      return {
        token: token ?? '',
        email: 'new.member@example.com',
        organization: 'Marina Group',
        role: 'Manager',
        invitedBy: 'Sara Khalil',
        expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 48).toISOString(),
        expired: false,
      };
    },
  });

export const useAcceptInvitation = () =>
  useMutation({
    mutationFn: async (_payload: { token: string; password?: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 500));
      return { ok: true } as const;
    },
  });
