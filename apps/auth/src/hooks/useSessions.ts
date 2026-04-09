import { useMutation, useQuery } from '@tanstack/react-query';
import type { SessionInfo } from '../components/SessionCard';

const MOCK_SESSIONS: SessionInfo[] = [
  {
    id: 'sess-1',
    device: 'MacBook Pro',
    browser: 'Chrome 130',
    os: 'macOS 14.4',
    location: 'Dubai, UAE',
    ip: '185.23.12.44',
    lastActive: 'Active now',
    current: true,
    type: 'desktop',
  },
  {
    id: 'sess-2',
    device: 'iPhone 15',
    browser: 'Safari',
    os: 'iOS 17.4',
    location: 'Dubai, UAE',
    ip: '94.200.x.x',
    lastActive: '2 hours ago',
    type: 'mobile',
  },
  {
    id: 'sess-3',
    device: 'Windows PC',
    browser: 'Edge',
    os: 'Windows 11',
    location: 'Riyadh, Saudi Arabia',
    ip: '37.34.x.x',
    lastActive: 'Yesterday',
    type: 'desktop',
  },
  {
    id: 'sess-4',
    device: 'iPad Pro',
    browser: 'Safari',
    os: 'iPadOS 17.4',
    location: 'Abu Dhabi, UAE',
    ip: '94.200.x.x',
    lastActive: '3 days ago',
    type: 'tablet',
  },
];

export const useSessions = () =>
  useQuery<SessionInfo[]>({
    queryKey: ['sessions'] as const,
    queryFn: async () => {
      await new Promise((resolve) => setTimeout(resolve, 300));
      return MOCK_SESSIONS;
    },
  });

export const useRevokeSession = () =>
  useMutation({
    mutationFn: async (_payload: { sessionId: string }) => {
      await new Promise((resolve) => setTimeout(resolve, 300));
      return { ok: true } as const;
    },
  });

export const useRevokeAllSessions = () =>
  useMutation({
    mutationFn: async () => {
      await new Promise((resolve) => setTimeout(resolve, 300));
      return { ok: true } as const;
    },
  });
