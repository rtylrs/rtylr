import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { securityStateSchema } from '../../schemas/security';
import { toast } from 'sonner';

export type { SecurityState, Session, TrustedDevice } from '../../schemas/security';

export const useSecurity = () => {
  return useQuery({
    queryKey: ['security'],
    queryFn: async () => {
      const res = await unwrap(safe.get<{ state?: unknown }>('/api/hub/security/v1/state'));
      return securityStateSchema.parse(res?.state ?? res);
    },
  });
};

export const useToggleTwoFactor = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (enabled: boolean) => {
      await unwrap(safe.patch('/api/hub/security/v1/2fa', { enabled }));
      return enabled;
    },
    onSuccess: (enabled) => {
      toast.success(enabled ? 'Two-factor enabled' : 'Two-factor disabled');
      qc.invalidateQueries({ queryKey: ['security'] });
    },
    onError: () => toast.error('Failed to update two-factor settings'),
  });
};

export const useRevokeSession = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (_id: string) => {
      await unwrap(safe.post('/api/hub/security/v1/invalidate-sessions'));
    },
    onSuccess: () => {
      toast.success('Signed out of all sessions');
      qc.invalidateQueries({ queryKey: ['security'] });
    },
    onError: () => toast.error('Failed to sign out sessions'),
  });
};
