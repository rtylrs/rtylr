import { useMutation, useQuery } from '@tanstack/react-query';

export interface TrustedDevice {
  id: string;
  name: string;
  lastUsed: string;
  browser?: string;
  os?: string;
  trustedSince: string;
}

const MOCK_DEVICES: TrustedDevice[] = [
  {
    id: 'dev-1',
    name: 'Personal MacBook',
    lastUsed: '5 minutes ago',
    browser: 'Chrome',
    os: 'macOS',
    trustedSince: '2 months ago',
  },
  {
    id: 'dev-2',
    name: 'Office iMac',
    lastUsed: 'Yesterday',
    browser: 'Safari',
    os: 'macOS',
    trustedSince: '4 months ago',
  },
  {
    id: 'dev-3',
    name: 'iPhone 15 Pro',
    lastUsed: '3 hours ago',
    browser: 'Safari Mobile',
    os: 'iOS',
    trustedSince: '1 month ago',
  },
];

export const useDevices = () =>
  useQuery<TrustedDevice[]>({
    queryKey: ['devices'] as const,
    queryFn: async () => {
      await new Promise((resolve) => setTimeout(resolve, 300));
      return MOCK_DEVICES;
    },
  });

export const useRemoveDevice = () =>
  useMutation({
    mutationFn: async (_payload: { deviceId: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 300));
      return { ok: true } as const;
    },
  });
