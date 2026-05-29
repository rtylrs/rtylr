import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

beforeEach(() => {
  vi.spyOn(document, 'cookie', 'get').mockReturnValue('');
  vi.spyOn(document, 'cookie', 'set').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('getCookie', () => {
  it('returns null when cookie does not exist', () => {
    vi.spyOn(document, 'cookie', 'get').mockReturnValue('');
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('returns the value of an existing cookie', () => {
    vi.spyOn(document, 'cookie', 'get').mockReturnValue('theme=dark; lang=en');
    expect(getCookie('theme')).toBe('dark');
  });

  it('returns null for empty document.cookie', () => {
    vi.spyOn(document, 'cookie', 'get').mockReturnValue('');
    expect(getCookie('anything')).toBeNull();
  });

  it('decodes URI-encoded cookie values', () => {
    vi.spyOn(document, 'cookie', 'get').mockReturnValue('name=hello%20world');
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
