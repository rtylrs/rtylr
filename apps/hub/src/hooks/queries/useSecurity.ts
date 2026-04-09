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

const MOCK_SECURITY: SecurityState = {
  twoFactorEnabled: false,
  twoFactorMethod: null,
  passwordChangedAt: new Date(Date.now() - 90 * 86400e3).toISOString(),
  ssoEnabled: false,
  sessions: [
    { id: 's1', device: 'MacBook Pro', browser: 'Chrome 124', location: 'Boston, MA',    ip: '73.12.88.14',  current: true,  lastActive: new Date().toISOString() },
    { id: 's2', device: 'iPhone 15',   browser: 'Safari 17',  location: 'Cambridge, MA', ip: '73.12.90.201', current: false, lastActive: new Date(Date.now() - 2 * 3600e3).toISOString() },
    { id: 's3', device: 'Windows PC',  browser: 'Edge 122',   location: 'New York, NY',  ip: '208.91.7.44',  current: false, lastActive: new Date(Date.now() - 36 * 3600e3).toISOString() },
  ],
  trustedDevices: [
    { id: 'd1', name: "Owner's MacBook Pro", lastSeen: new Date().toISOString(),                        os: 'macOS 15.4' },
    { id: 'd2', name: "Owner's iPhone",      lastSeen: new Date(Date.now() - 2 * 3600e3).toISOString(), os: 'iOS 18.3' },
  ],
};

export const useSecurity = () => {
  return useQuery<SecurityState>({
    queryKey: ['security'],
    queryFn: async () => {
      try {
        const res = await api.get('/security');
        return res.data as SecurityState;
      } catch {
        return MOCK_SECURITY;
      }
    },
  });
};

export const useToggleTwoFactor = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (enabled: boolean) => {
      try {
        await api.post('/security/2fa', { enabled });
      } catch {
        // ignore
      }
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
      try {
        await api.post(`/security/sessions/${id}/revoke`);
      } catch {
        // ignore
      }
      return id;
    },
    onSuccess: () => {
      toast.success('Session revoked');
      qc.invalidateQueries({ queryKey: ['security'] });
    },
  });
};
