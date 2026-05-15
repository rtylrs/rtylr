import { RealtimeClient } from '@voxire/api';

function getTokenFromCookie(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

function resolveWsUrl(): string {
  const base = (import.meta.env['VITE_API_URL'] as string | undefined) ?? 'https://api.voxire.com/';
  const stripped = base.replace(/\/$/, '');
  const wsScheme = stripped.startsWith('https://') ? 'wss://' : 'ws://';
  return `${wsScheme}${stripped.replace(/^https?:\/\//, '')}/pos/ws`;
}

export const posRealtime = new RealtimeClient({
  url: resolveWsUrl(),
  getToken: getTokenFromCookie,
});
