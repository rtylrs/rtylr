import { afterEach, beforeEach, describe, expect, test } from 'vitest';

import { getCookie, getDefaultCookieDomain, removeCookie, setCookie } from './cookies';

/**
 * Unit tests for the cookie helpers. The most important regression
 * here is from commit c81034fc: removeCookie MUST emit the Domain
 * attribute when given one, and the canonical logout pattern is two
 * calls (cross-domain + host-only). A plain `document.cookie =
 * 'token=; Path=/'` without Domain leaves the parent-domain cookie
 * alive on .rtylr.com, and auth.rtylr.com bounces the user right
 * back in.
 *
 * happy-dom doesn't enforce the Domain= cross-host semantics that bit
 * us in production (localhost cookies are host-only either way), so
 * the assertions here are observational: cookie present → call(s) →
 * cookie gone. The string-level Set-Cookie format that real browsers
 * actually validate is exercised by the e2e logout spec
 * (`e2e/auth-logout.spec.ts`) running in Chromium.
 */

function clearAllCookies(): void {
  document.cookie.split(';').forEach((c) => {
    const name = c.split('=')[0]?.trim();
    if (name) document.cookie = `${name}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Path=/`;
  });
}

describe('cookies helpers', () => {
  beforeEach(() => clearAllCookies());
  afterEach(() => clearAllCookies());

  test('setCookie + getCookie round-trip', () => {
    setCookie('token', 'abc');
    expect(getCookie('token')).toBe('abc');
  });

  test('getCookie returns null for absent cookie', () => {
    expect(getCookie('does-not-exist')).toBeNull();
  });

  test('removeCookie clears a host-only cookie', () => {
    setCookie('token', 'abc');
    expect(getCookie('token')).toBe('abc');
    removeCookie('token');
    // happy-dom may retain a zero-length value rather than purging
    // the cookie outright; either is "cleared" for our purposes.
    expect(getCookie('token') || null).toBeNull();
  });

  test('removeCookie with no domain does not throw', () => {
    setCookie('token', 'abc');
    expect(() => removeCookie('token')).not.toThrow();
  });

  test('removeCookie with explicit domain does not throw on localhost', () => {
    // The shape that bit us in production: removeCookie('token', { domain: getDefaultCookieDomain() }).
    // On localhost getDefaultCookieDomain() returns undefined; in
    // production it returns '.rtylr.com'. We pass a literal domain
    // here to verify the helper accepts it without erroring.
    setCookie('token', 'abc');
    expect(() => removeCookie('token', { domain: '.rtylr.com' })).not.toThrow();
  });

  test('the documented logout pattern (cross-domain + host-only) clears the cookie', () => {
    // Belt + suspenders: every logout site does both calls. If either
    // is removed, the e2e logout spec will fail in Chromium where
    // Domain= semantics are enforced. Here we just verify the
    // local-observable end state - cookie is gone after both calls.
    setCookie('token', 'fake-jwt');
    expect(getCookie('token')).toBe('fake-jwt');
    removeCookie('token', { domain: getDefaultCookieDomain() });
    removeCookie('token');
    expect(getCookie('token') || null).toBeNull();
  });

  test('getDefaultCookieDomain returns undefined on localhost (host-only)', () => {
    // happy-dom defaults the test page to `localhost`. Setting a
    // cookie with `Domain=.rtylr.com` from localhost would be
    // rejected by real browsers; we don't want the helper to ever
    // hand that out. undefined → setCookie omits Domain → host-only.
    expect(getDefaultCookieDomain()).toBeUndefined();
  });
});
