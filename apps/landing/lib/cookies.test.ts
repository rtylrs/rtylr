import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

let fakeCookie = '';

beforeEach(() => {
  fakeCookie = '';
  Object.defineProperty(document, 'cookie', {
    configurable: true,
    get: () => fakeCookie,
    set: (val: string) => {
      fakeCookie = val;
    },
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('getCookie', () => {
  it('returns null when cookie does not exist', () => {
    fakeCookie = '';
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('returns the value of an existing cookie', () => {
    fakeCookie = 'theme=dark; lang=en';
    expect(getCookie('theme')).toBe('dark');
  });

  it('returns null for empty document.cookie', () => {
    fakeCookie = '';
    expect(getCookie('anything')).toBeNull();
  });

  it('decodes URI-encoded cookie values', () => {
    fakeCookie = 'name=hello%20world';
    expect(getCookie('name')).toBe('hello world');
  });
});

describe('setCookie', () => {
  it('sets a basic cookie without options', () => {
    setCookie('token', 'abc123');
    expect(fakeCookie).toContain('token=abc123');
  });

  it('includes Expires when days is provided', () => {
    setCookie('token', 'abc123', { days: 7 });
    expect(fakeCookie).toContain('Expires=');
  });

  it('includes Domain when provided', () => {
    setCookie('token', 'abc123', { domain: '.rtylr.com' });
    expect(fakeCookie).toContain('Domain=.rtylr.com');
  });

  it('includes SameSite', () => {
    setCookie('x', 'y', { sameSite: 'strict' });
    expect(fakeCookie).toContain('SameSite=strict');
  });
});

describe('removeCookie', () => {
  it('sets cookie with expired date to remove it', () => {
    removeCookie('token');
    expect(fakeCookie).toContain('Expires=Thu, 01 Jan 1970');
  });

  it('includes Domain when provided', () => {
    removeCookie('token', { domain: '.rtylr.com' });
    expect(fakeCookie).toContain('Domain=.rtylr.com');
  });
});

describe('getDefaultCookieDomain', () => {
  it('returns undefined for localhost', () => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      writable: true,
      value: { hostname: 'localhost' },
    });
    expect(getDefaultCookieDomain()).toBeUndefined();
  });

  it('returns .rtylr.com for rtylr.com hostname', () => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      writable: true,
      value: { hostname: 'rtylr.com' },
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });

  it('returns .rtylr.com for subdomain of rtylr.com', () => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      writable: true,
      value: { hostname: 'app.rtylr.com' },
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });
});
