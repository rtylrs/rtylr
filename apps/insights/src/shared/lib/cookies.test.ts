import { describe, it, expect, beforeEach } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

describe('getCookie', () => {
  beforeEach(() => {
    document.cookie.split(';').forEach((c) => {
      const name = c.trim().split('=')[0];
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/`;
    });
  });

  it('returns null when the cookie is not set', () => {
    expect(getCookie('missing')).toBeNull();
  });

  it('returns the cookie value when it is set', () => {
    document.cookie = 'testKey=hello';
    expect(getCookie('testKey')).toBe('hello');
  });

  it('decodes URI-encoded cookie values', () => {
    document.cookie = `encoded=${encodeURIComponent('hello world')}`;
    expect(getCookie('encoded')).toBe('hello world');
  });

  it('returns null for a name that does not match', () => {
    document.cookie = 'a=1';
    expect(getCookie('b')).toBeNull();
  });
});

describe('setCookie', () => {
  beforeEach(() => {
    document.cookie.split(';').forEach((c) => {
      const name = c.trim().split('=')[0];
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/`;
    });
  });

  it('sets a cookie that can be read back', () => {
    setCookie('myKey', 'myValue', { days: 7 });
    expect(getCookie('myKey')).toBe('myValue');
  });

  it('encodes special characters in values', () => {
    setCookie('special', 'a b&c');
    expect(getCookie('special')).toBe('a b&c');
  });
});

describe('removeCookie', () => {
  it('removes a previously set cookie', () => {
    setCookie('toRemove', 'gone');
    removeCookie('toRemove');
    // happy-dom may return '' or null after expiry-based removal
    const result = getCookie('toRemove');
    expect(result === null || result === '').toBe(true);
  });
});

describe('getDefaultCookieDomain', () => {
  it('returns undefined on localhost', () => {
    Object.defineProperty(window, 'location', {
      value: { hostname: 'localhost' },
      writable: true,
    });
    expect(getDefaultCookieDomain()).toBeUndefined();
  });

  it('returns ".rtylr.com" on the production domain', () => {
    Object.defineProperty(window, 'location', {
      value: { hostname: 'app.rtylr.com' },
      writable: true,
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });

  it('returns ".rtylr.com" on the apex domain', () => {
    Object.defineProperty(window, 'location', {
      value: { hostname: 'rtylr.com' },
      writable: true,
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });
});
