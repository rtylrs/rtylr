import { resolveApiBase, resolveAuthUrl } from '@voxire/config';
import { removeCookie, getDefaultCookieDomain } from './cookies';
import { useUserStore } from './store';
import type { AppId } from './types';

/**
 * Clears the SSO token cookie (host + parent domain), resets the shared user store,
 * and redirects to auth with ?redirect= so the user can sign in again for this app.
 *
 * `resolveAuthUrl()` returns the hosted sign-in application URL.
 */
export function signOut(appId: AppId): void {
  const refreshToken = document.cookie.match(/(?:^|;\s*)refresh_token=([^;]+)/)?.[1];
  if (refreshToken) {
    try {
      void fetch(`${resolveApiBase().replace(/\/+$/, '')}/common/auth/v1/revoke`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: decodeURIComponent(refreshToken) }),
        keepalive: true,
      }).catch(() => undefined);
    } catch {
      // Clearing local credentials remains the priority if the cookie is malformed.
    }
  }
  removeCookie('token', { domain: getDefaultCookieDomain() });
  removeCookie('token');
  removeCookie('refresh_token', { domain: getDefaultCookieDomain() });
  removeCookie('refresh_token');
  useUserStore.getState().logout();
  const u = new URL(resolveAuthUrl());
  u.searchParams.set('redirect', appId);
  window.location.href = u.toString();
}
