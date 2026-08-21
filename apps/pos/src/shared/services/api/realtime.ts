import { RealtimeClient } from '@voxire/api';
import { resolveWsBaseUrl } from '../server/serverConfig';
import { getCookie } from './cookies';

const getTokenFromCookie = (): string | null => getCookie('token');

function resolveWsUrl(): string {
  // Base resolves self-host → VITE_API_URL → hosted default.
  return `${resolveWsBaseUrl()}/pos/ws`;
}

export const posRealtime = new RealtimeClient({
  url: resolveWsUrl(),
  getToken: getTokenFromCookie,
});
