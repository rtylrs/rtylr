import { describe, it, expect } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

describe('getCookie', () => {
  it('returns null when cookie does not exist', () => {
    expect(getCookie('definitely_missing_cookie_xyz123')).toBeNull();
  });
  it('returns a cookie value after setting it', () => {
    setCookie('testkey_x', 'testval_x');
    expect(getCookie('testkey_x')).toBe('testval_x');
  });
});

describe('getDefaultCookieDomain', () => {
  it('returns .rtylr.com for rtylr.com', () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { hostname: 'rtylr.com' },
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });
  it('returns .rtylr.com for subdomains', () => {
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { hostname: 'dash.rtylr.com' },
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });
});
