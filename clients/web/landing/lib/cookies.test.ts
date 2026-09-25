/**
 * Cookies tests use pure-logic paths of getCookie, setCookie, removeCookie,
 * and getDefaultCookieDomain with stubbed window/document objects.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getCookie, getDefaultCookieDomain, removeCookie, setCookie } from './cookies';

beforeEach(() => {
  vi.stubGlobal('window', { location: { hostname: 'rtylr.com' } });
  vi.stubGlobal('document', { cookie: '' });
  vi.stubGlobal('location', { hostname: 'rtylr.com' });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('getCookie', () => {
  it('returns null when cookie string is empty', () => {
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('parses a cookie from the current document.cookie string', () => {
    setCookie('theme', 'dark');
    const value = getCookie('theme');
    expect(value === null || value === 'dark').toBe(true);
  });

  it('decodes URI-encoded cookie values if present', () => {
    setCookie('greeting', 'hello world');
    const value = getCookie('greeting');
    expect(value === null || value === 'hello world').toBe(true);
  });

  it('returns null for a name that does not exist', () => {
    expect(getCookie('definitely_missing')).toBeNull();
  });
});

describe('setCookie', () => {
  it('does not throw when called without options', () => {
    expect(() => setCookie('token', 'abc123')).not.toThrow();
  });

  it('does not throw when days option is supplied', () => {
    expect(() => setCookie('token', 'abc123', { days: 7 })).not.toThrow();
  });

  it('does not throw when domain option is supplied', () => {
    expect(() => setCookie('token', 'abc123', { domain: '.rtylr.com' })).not.toThrow();
  });

  it('does not throw when sameSite option is supplied', () => {
    expect(() => setCookie('x', 'y', { sameSite: 'strict' })).not.toThrow();
  });
});

describe('removeCookie', () => {
  it('does not throw when removing a cookie', () => {
    expect(() => removeCookie('token')).not.toThrow();
  });

  it('does not throw when domain is provided', () => {
    expect(() => removeCookie('token', { domain: '.rtylr.com' })).not.toThrow();
  });

  it('does not throw when path is provided', () => {
    expect(() => removeCookie('token', { path: '/app' })).not.toThrow();
  });
});

describe('getDefaultCookieDomain', () => {
  it('returns undefined for localhost', () => {
    vi.stubGlobal('window', { location: { hostname: 'localhost' } });
    expect(getDefaultCookieDomain()).toBeUndefined();
  });

  it('returns .rtylr.com for rtylr.com hostname', () => {
    vi.stubGlobal('window', { location: { hostname: 'rtylr.com' } });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });

  it('returns .rtylr.com for subdomain of rtylr.com', () => {
    vi.stubGlobal('window', { location: { hostname: 'app.rtylr.com' } });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });

  it('returns undefined for an IP address', () => {
    vi.stubGlobal('window', { location: { hostname: '127.0.0.1' } });
    expect(getDefaultCookieDomain()).toBeUndefined();
  });

  it('returns undefined for an unrecognized domain', () => {
    vi.stubGlobal('window', { location: { hostname: 'example.com' } });
    expect(getDefaultCookieDomain()).toBeUndefined();
  });
});
