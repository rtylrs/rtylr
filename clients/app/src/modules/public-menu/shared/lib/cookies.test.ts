import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { getCookie, setCookie, removeCookie, getDefaultCookieDomain } from './cookies';

describe('cookies helper', () => {
  beforeEach(() => {
    // Clear cookies before each test
    document.cookie.split(';').forEach((c) => {
      const eqPos = c.indexOf('=');
      const name = eqPos > -1 ? c.substring(0, eqPos).trim() : c.trim();
      if (name) {
        document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;max-age=0`;
      }
    });
  });

  describe('getCookie & setCookie', () => {
    it('sets and retrieves a simple cookie', () => {
      setCookie('language', 'ar');
      expect(getCookie('language')).toBe('ar');
    });

    it('returns null when cookie does not exist', () => {
      expect(getCookie('nonexistent')).toBeNull();
    });

    it('encodes and decodes special characters', () => {
      setCookie('test_special', 'hello world = & %');
      expect(getCookie('test_special')).toBe('hello world = & %');
    });

    it('sets days expiry option', () => {
      setCookie('expires_test', 'val', { days: 7 });
      expect(getCookie('expires_test')).toBe('val');
    });

    it('sets path and sameSite options', () => {
      setCookie('options_test', 'val', { path: '/', sameSite: 'strict' });
      expect(getCookie('options_test')).toBe('val');
    });
  });

  describe('removeCookie', () => {
    it('removes an existing cookie', () => {
      setCookie('to_remove', 'active');
      expect(getCookie('to_remove')).toBe('active');

      removeCookie('to_remove');
      expect(getCookie('to_remove')).toBeNull();
    });
  });

  describe('getDefaultCookieDomain', () => {
    const originalLocation = window.location;

    afterEach(() => {
      Object.defineProperty(window, 'location', {
        writable: true,
        value: originalLocation,
      });
    });

    it('returns undefined for localhost', () => {
      Object.defineProperty(window, 'location', {
        writable: true,
        value: { ...originalLocation, hostname: 'localhost' },
      });
      expect(getDefaultCookieDomain()).toBeUndefined();
    });

    it('returns undefined for numeric IP addresses', () => {
      Object.defineProperty(window, 'location', {
        writable: true,
        value: { ...originalLocation, hostname: '127.0.0.1' },
      });
      expect(getDefaultCookieDomain()).toBeUndefined();
    });

    it('returns .rtylr.com for rtylr.com domains', () => {
      Object.defineProperty(window, 'location', {
        writable: true,
        value: { ...originalLocation, hostname: 'rtylr.com' },
      });
      expect(getDefaultCookieDomain()).toBe('.rtylr.com');

      Object.defineProperty(window, 'location', {
        writable: true,
        value: { ...originalLocation, hostname: 'app.rtylr.com' },
      });
      expect(getDefaultCookieDomain()).toBe('.rtylr.com');
    });

    it('returns undefined for unrelated domains', () => {
      Object.defineProperty(window, 'location', {
        writable: true,
        value: { ...originalLocation, hostname: 'otherdomain.org' },
      });
      expect(getDefaultCookieDomain()).toBeUndefined();
    });
  });
});
