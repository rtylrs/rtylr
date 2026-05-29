import { describe, it, expect, vi, afterEach } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

let cookieStore = '';

function mockCookie() {
  const proto = Object.getPrototypeOf(document);
  Object.defineProperty(proto, 'cookie', {
    configurable: true,
    get() {
      return cookieStore;
    },
    set(val: string) {
      cookieStore = val;
    },
  });
}

mockCookie();

afterEach(() => {
  cookieStore = '';
  vi.restoreAllMocks();
});

describe('getCookie', () => {
  it('returns null when cookie does not exist', () => {
    cookieStore = '';
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('returns the value of an existing cookie', () => {
    cookieStore = 'theme=dark; lang=en';
    expect(getCookie('theme')).toBe('dark');
  });

  it('returns null for empty document.cookie', () => {
    cookieStore = '';
    expect(getCookie('anything')).toBeNull();
  });

  it('decodes URI-encoded cookie values', () => {
    cookieStore = 'name=hello%20world';
    expect(getCookie('name')).toBe('hello world');
  });
});

describe('setCookie', () => {
  it('sets a basic cookie without options', () => {
    setCookie('token', 'abc123');
    expect(cookieStore).toContain('token=abc123');
  });

  it('includes Expires when days is provided', () => {
    setCookie('token', 'abc123', { days: 7 });
    expect(cookieStore).toContain('Expires=');
  });

  it('includes Domain when provided', () => {
    setCookie('token', 'abc123', { domain: '.rtylr.com' });
    expect(cookieStore).toContain('Domain=.rtylr.com');
  });

  it('includes SameSite', () => {
    setCookie('x', 'y', { sameSite: 'strict' });
    expect(cookieStore).toContain('SameSite=strict');
  });
});

describe('removeCookie', () => {
  it('sets cookie with expired date to remove it', () => {
    removeCookie('token');
    expect(cookieStore).toContain('Expires=Thu, 01 Jan 1970');
  });

  it('includes Domain when provided', () => {
    removeCookie('token', { domain: '.rtylr.com' });
    expect(cookieStore).toContain('Domain=.rtylr.com');
  });
});

describe('getDefaultCookieDomain', () => {
  it('returns undefined for localhost', () => {
    vi.stubGlobal('location', { hostname: 'localhost' });
    expect(getDefaultCookieDomain()).toBeUndefined();
  });

  it('returns .rtylr.com for rtylr.com hostname', () => {
    vi.stubGlobal('location', { hostname: 'rtylr.com' });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });

  it('returns .rtylr.com for subdomain of rtylr.com', () => {
    vi.stubGlobal('location', { hostname: 'app.rtylr.com' });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });
});
