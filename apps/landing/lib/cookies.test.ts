import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

// Use a writable property on document.prototype to intercept cookie access
const cookies: Map<string, string> = new Map();
let rawCookie = '';

// Override document.cookie via prototype in a way happy-dom supports
const DocumentProto = Object.getPrototypeOf(document);
const origDescriptor = Object.getOwnPropertyDescriptor(DocumentProto, 'cookie');

Object.defineProperty(DocumentProto, 'cookie', {
  configurable: true,
  get(): string {
    return rawCookie;
  },
  set(val: string): void {
    rawCookie = val;
  },
});

beforeEach(() => {
  rawCookie = '';
});

describe('getCookie', () => {
  it('returns null when cookie does not exist', () => {
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('returns the value of an existing cookie', () => {
    rawCookie = 'theme=dark; lang=en';
    expect(getCookie('theme')).toBe('dark');
  });

  it('returns null for empty document.cookie', () => {
    rawCookie = '';
    expect(getCookie('anything')).toBeNull();
  });

  it('decodes URI-encoded cookie values', () => {
    rawCookie = 'name=hello%20world';
    expect(getCookie('name')).toBe('hello world');
  });
});

describe('setCookie', () => {
  it('sets a basic cookie without options', () => {
    setCookie('token', 'abc123');
    expect(rawCookie).toContain('token=abc123');
  });

  it('includes Expires when days is provided', () => {
    setCookie('token', 'abc123', { days: 7 });
    expect(rawCookie).toContain('Expires=');
  });

  it('includes Domain when provided', () => {
    setCookie('token', 'abc123', { domain: '.rtylr.com' });
    expect(rawCookie).toContain('Domain=.rtylr.com');
  });

  it('includes SameSite', () => {
    setCookie('x', 'y', { sameSite: 'strict' });
    expect(rawCookie).toContain('SameSite=strict');
  });
});

describe('removeCookie', () => {
  it('sets cookie with expired date to remove it', () => {
    removeCookie('token');
    expect(rawCookie).toContain('Expires=Thu, 01 Jan 1970');
  });

  it('includes Domain when provided', () => {
    removeCookie('token', { domain: '.rtylr.com' });
    expect(rawCookie).toContain('Domain=.rtylr.com');
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
