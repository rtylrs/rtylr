export const HOSTED_API_URL = 'https://api.voxire.com/';
export const HOSTED_AUTH_URL = 'https://sso.rtylr.com';

const configuredApiUrl = (): string =>
  (import.meta as ImportMeta & { env: Record<string, string | undefined> }).env.VITE_API_URL ||
  HOSTED_API_URL;

export function resolveApiBaseUrl(): string {
  return `${configuredApiUrl().trim().replace(/\/+$/, '')}/`;
}

export function resolveWsBaseUrl(): string {
  const base = resolveApiBaseUrl().replace(/\/$/, '');
  const scheme = base.startsWith('https://') ? 'wss://' : 'ws://';
  return `${scheme}${base.replace(/^https?:\/\//, '')}`;
}

export function resolveAuthUrl(): string {
  return HOSTED_AUTH_URL;
}
