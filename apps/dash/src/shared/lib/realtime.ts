import { useEffect } from 'react';
import { getRealtimeClient, type RealtimeEvent } from '@voxire/api';

function getTokenFromCookie(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

function resolveWsUrl(): string {
  const base = (import.meta.env.VITE_API_URL as string | undefined) ?? 'https://api.voxire.com/';
  const stripped = base.replace(/\/$/, '');
  const wsScheme = stripped.startsWith('https://') ? 'wss://' : 'ws://';
  return `${wsScheme}${stripped.replace(/^https?:\/\//, '')}/dash/ws`;
}

const client = getRealtimeClient({
  url: resolveWsUrl(),
  getToken: getTokenFromCookie,
});

export function useRealtime(topics: string[], handler: (event: RealtimeEvent) => void): void {
  useEffect(() => {
    if (topics.length === 0) return;
    const unsubs = topics.map((t) => client.subscribe(t, handler));
    return () => {
      unsubs.forEach((u) => u());
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [topics.join('|'), handler]);
}
