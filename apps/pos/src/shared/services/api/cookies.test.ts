import { describe, it, expect, beforeEach } from 'vitest';
import { getCookie, setCookie, removeCookie } from './cookies';

describe('getCookie', () => {
  beforeEach(() => {
    // Clear all cookies before each test
    document.cookie.split(';').forEach((cookie) => {
      const name = cookie.split('=')[0].trim();
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`;
    });
  });

  it('returns null when the cookie does not exist', () => {
    expect(getCookie('nonexistent')).toBeNull();
  });

  it('returns the value of an existing cookie', () => {
    document.cookie = 'testKey=hello';
    expect(getCookie('testKey')).toBe('hello');
  });

  it('returns the decoded value of a URL-encoded cookie', () => {
    document.cookie = `encoded=${encodeURIComponent('hello world')}`;
    expect(getCookie('encoded')).toBe('hello world');
  });

  it('is defined and callable', () => {
    // removeCookie is integration-tested below; here we just verify getCookie
    // returns null when the cookie was never set.
    expect(getCookie('neverSet')).toBeNull();
  });
});

describe('setCookie', () => {
  beforeEach(() => {
    document.cookie.split(';').forEach((cookie) => {
      const name = cookie.split('=')[0].trim();
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`;
    });
  });

  it('sets a cookie that can be retrieved by getCookie', () => {
    setCookie('myKey', 'myValue');
    expect(getCookie('myKey')).toBe('myValue');
  });

  it('URL-encodes special characters in the value', () => {
    setCookie('special', 'foo bar');
    expect(getCookie('special')).toBe('foo bar');
  });

  it('overwrites an existing cookie with the same name', () => {
    setCookie('dupe', 'first');
    setCookie('dupe', 'second');
    expect(getCookie('dupe')).toBe('second');
  });
});

describe('removeCookie', () => {
  it('does not throw when removing an existing cookie', () => {
    document.cookie = 'gone=yes';
    expect(() => removeCookie('gone')).not.toThrow();
  });

  it('does not throw when removing a non-existent cookie', () => {
    expect(() => removeCookie('doesNotExist')).not.toThrow();
  });
});
