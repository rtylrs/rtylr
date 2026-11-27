import type { AppId } from '@voxire/types';
import { APP_URLS } from './apps';
import { isTenantMenuHost } from './publicMenu';

const env = (import.meta as ImportMeta & { env?: Record<string, string | undefined> }).env;

function stripTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '');
}

export function resolveAppUrl(appId: AppId): string {
  return APP_URLS[appId];
}

export function resolveAuthUrl(): string {
  return APP_URLS.auth;
}

export function resolveApiBase(): string {
  return stripTrailingSlash(env?.VITE_API_URL || APP_URLS.api);
}

export function resolveCdnBaseUrl(): string {
  return stripTrailingSlash(env?.VITE_CDN_URL || APP_URLS.api);
}

export function resolveCookieDomain(): string | undefined {
  if (typeof window === 'undefined') return undefined;
  const hostname = window.location.hostname.toLowerCase();
  if (hostname === 'localhost' || /^\d+\.\d+\.\d+\.\d+$/.test(hostname)) return undefined;
  // Never write .rtylr.com-scoped cookies from public tenant storefront hosts.
  if (isTenantMenuHost(hostname)) return undefined;
  if (hostname === 'rtylr.com' || hostname.endsWith('.rtylr.com')) return '.rtylr.com';
  return undefined;
}
