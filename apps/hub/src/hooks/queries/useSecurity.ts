import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { toast } from 'sonner';

export interface Session {
  id: string;
  device: string;
  browser: string;
  location: string;
  ip: string;
  current: boolean;
  lastActive: string;
}

export interface TrustedDevice {
  id: string;
  name: string;
  lastSeen: string;
  os: string;
}

export interface SecurityState {
  twoFactorEnabled: boolean;
  twoFactorMethod: 'app' | 'sms' | null;
  sessions: Session[];
  trustedDevices: TrustedDevice[];
  passwordChangedAt?: string;
  ssoEnabled: boolean;
}

export const useSecurity = () => {
  return useQuery<SecurityState>({
    queryKey: ['security'],
    queryFn: async () => {
      const res = await api.get('/security');
      return res.data as SecurityState;
    },
  });
};

export const useToggleTwoFactor = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (enabled: boolean) => {
      await api.post('/security/2fa', { enabled });
      return enabled;
    },
    onSuccess: (enabled) => {
      toast.success(enabled ? 'Two-factor enabled' : 'Two-factor disabled');
      qc.invalidateQueries({ queryKey: ['security'] });
    },
  });
};

export const useRevokeSession = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/security/sessions/${id}/revoke`);
      return id;
    },
    onSuccess: () => {
      toast.success('Session revoked');
      qc.invalidateQueries({ queryKey: ['security'] });
    },
  });
};
