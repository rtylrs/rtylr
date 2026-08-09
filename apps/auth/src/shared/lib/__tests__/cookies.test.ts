import { describe, it, expect, beforeEach, vi } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from '../cookies';

beforeEach(() => {
  // Reset document.cookie between tests
  document.cookie.split('; ').forEach((c) => {
    const name = c.split('=')[0];
    if (name) document.cookie = `${name}=; Max-Age=0; Path=/`;
  });
});

describe('getCookie', () => {
  it('returns null when the cookie is absent', () => {
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('reads a cookie that was just set', () => {
    document.cookie = 'mykey=myvalue; Path=/';
    expect(getCookie('mykey')).toBe('myvalue');
  });

  it('decodes URI-encoded values', () => {
    document.cookie = `encoded=${encodeURIComponent('hello world')}; Path=/`;
    expect(getCookie('encoded')).toBe('hello world');
  });
});

describe('setCookie', () => {
  it('sets a cookie readable by getCookie', () => {
    setCookie('testkey', 'testval');
    expect(getCookie('testkey')).toBe('testval');
  });

  it('encodes special characters in the value', () => {
    setCookie('special', 'a b+c');
    const raw = document.cookie;
    expect(raw).toContain('special=');
    expect(getCookie('special')).toBe('a b+c');
  });
});

describe('removeCookie', () => {
  it('removes a previously set cookie', () => {
    setCookie('toremove', 'value');
    removeCookie('toremove');
    const result = getCookie('toremove');
    // happy-dom may return '' instead of null for an expired cookie
    expect(result === null || result === '').toBe(true);
  });
});

describe('getDefaultCookieDomain', () => {
  // Note: isBrowser is cached at module load time, so stubbing window to
  // undefined after load does not change the cached value. The no-window
  // branch is tested by the SSR build; here we only test the browser paths.

  it('returns undefined for localhost', () => {
    vi.stubGlobal('window', { location: { hostname: 'localhost' } });
    expect(getDefaultCookieDomain()).toBeUndefined();
    vi.unstubAllGlobals();
  });

  it('returns ".rtylr.com" for an rtylr subdomain', () => {
    vi.stubGlobal('window', { location: { hostname: 'auth.rtylr.com' } });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
    vi.unstubAllGlobals();
  });

  it('returns ".rtylr.com" for the apex domain', () => {
    vi.stubGlobal('window', { location: { hostname: 'rtylr.com' } });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
    vi.unstubAllGlobals();
  });
});
