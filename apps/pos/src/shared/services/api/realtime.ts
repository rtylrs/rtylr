import { RealtimeClient } from '@voxire/api';
import { resolveWsBaseUrl } from '../server/serverConfig';

function getTokenFromCookie(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

function resolveWsUrl(): string {
  // Base resolves self-host → VITE_API_URL → hosted default.
  return `${resolveWsBaseUrl()}/pos/ws`;
}

export const posRealtime = new RealtimeClient({
  url: resolveWsUrl(),
  getToken: getTokenFromCookie,
});
