import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

beforeEach(() => {
  // Reset document.cookie before each test
  Object.defineProperty(document, 'cookie', {
    writable: true,
    value: '',
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('getCookie', () => {
  it('returns null when cookie does not exist', () => {
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('returns the value of an existing cookie', () => {
    Object.defineProperty(document, 'cookie', {
      writable: true,
      value: 'theme=dark; lang=en',
    });
    expect(getCookie('theme')).toBe('dark');
  });

  it('returns null for empty document.cookie', () => {
    Object.defineProperty(document, 'cookie', {
      writable: true,
      value: '',
    });
    expect(getCookie('anything')).toBeNull();
  });

  it('decodes URI-encoded cookie values', () => {
    Object.defineProperty(document, 'cookie', {
      writable: true,
      value: 'name=hello%20world',
    });
    expect(getCookie('name')).toBe('hello world');
  });
});

describe('setCookie', () => {
  it('sets a basic cookie without options', () => {
    const spy = vi.spyOn(document, 'cookie', 'set');
    setCookie('token', 'abc123');
    expect(spy).toHaveBeenCalled();
  });

  it('includes Expires when days is provided', () => {
    const spy = vi.spyOn(document, 'cookie', 'set');
    setCookie('token', 'abc123', { days: 7 });
    expect(spy.mock.calls[0][0]).toContain('Expires=');
  });

  it('includes Domain when provided', () => {
    const spy = vi.spyOn(document, 'cookie', 'set');
    setCookie('token', 'abc123', { domain: '.rtylr.com' });
    expect(spy.mock.calls[0][0]).toContain('Domain=.rtylr.com');
  });

  it('includes SameSite', () => {
    const spy = vi.spyOn(document, 'cookie', 'set');
    setCookie('x', 'y', { sameSite: 'strict' });
    expect(spy.mock.calls[0][0]).toContain('SameSite=strict');
  });
});

describe('removeCookie', () => {
  it('sets cookie with expired date to remove it', () => {
    const spy = vi.spyOn(document, 'cookie', 'set');
    removeCookie('token');
    expect(spy.mock.calls[0][0]).toContain('Expires=Thu, 01 Jan 1970');
  });

  it('includes Domain when provided', () => {
    const spy = vi.spyOn(document, 'cookie', 'set');
    removeCookie('token', { domain: '.rtylr.com' });
    expect(spy.mock.calls[0][0]).toContain('Domain=.rtylr.com');
  });
});

describe('getDefaultCookieDomain', () => {
  it('returns undefined for localhost', () => {
    Object.defineProperty(window, 'location', {
      value: { hostname: 'localhost' },
      writable: true,
    });
    expect(getDefaultCookieDomain()).toBeUndefined();
  });

  it('returns .rtylr.com for rtylr.com hostname', () => {
    Object.defineProperty(window, 'location', {
      value: { hostname: 'rtylr.com' },
      writable: true,
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });

  it('returns .rtylr.com for subdomain of rtylr.com', () => {
    Object.defineProperty(window, 'location', {
      value: { hostname: 'app.rtylr.com' },
      writable: true,
    });
    expect(getDefaultCookieDomain()).toBe('.rtylr.com');
  });
});
