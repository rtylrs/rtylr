import { resolveAuthUrl } from '@voxire/config';
import { removeCookie, getDefaultCookieDomain } from './cookies';
import { useUserStore } from './store';
import type { AppId } from './types';

/**
 * Clears the SSO token cookie (host + parent domain), resets the shared user store,
 * and redirects to auth with ?redirect= so the user can sign in again for this app.
 *
 * `resolveAuthUrl()` returns the runtime auth URL in self-host deployments and
 * the static `sso.rtylr.com` default everywhere else.
 */
export function signOut(appId: AppId): void {
  removeCookie('token', { domain: getDefaultCookieDomain() });
  removeCookie('token');
  useUserStore.getState().logout();
  const u = new URL(resolveAuthUrl());
  u.searchParams.set('redirect', appId);
  window.location.href = u.toString();
}
