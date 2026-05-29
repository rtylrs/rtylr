import { describe, it, expect } from 'vitest';
import {
  joinPath,
  lastPathSegment,
  stripQueryAndHash,
  extractQueryString,
  isActivePath,
  withQueryParams,
  isAbsoluteUrl,
  getOrigin,
} from './urlUtils';

describe('joinPath', () => {
  it('joins segments without double slashes', () => {
    expect(joinPath('/app/', '/dashboard/', 'settings')).toBe('app/dashboard/settings');
  });

  it('handles empty segments', () => {
    expect(joinPath('a', '', 'b')).toBe('a/b');
  });
});

describe('lastPathSegment', () => {
  it('returns last segment', () =>
    expect(lastPathSegment('/app/settings/profile')).toBe('profile'));
  it('handles trailing slash', () => expect(lastPathSegment('/app/settings/')).toBe('settings'));
});

describe('stripQueryAndHash', () => {
  it('removes query string', () => expect(stripQueryAndHash('/page?q=1')).toBe('/page'));
  it('removes hash', () => expect(stripQueryAndHash('/page#section')).toBe('/page'));
  it('removes both', () => expect(stripQueryAndHash('/page?q=1#section')).toBe('/page'));
});

describe('extractQueryString', () => {
  it('returns query string without ?', () =>
    expect(extractQueryString('/page?a=1&b=2')).toBe('a=1&b=2'));
  it('returns empty for no query', () => expect(extractQueryString('/page')).toBe(''));
  it('strips hash from query', () => expect(extractQueryString('/page?q=1#anchor')).toBe('q=1'));
});

describe('isActivePath', () => {
  it('exact match', () => expect(isActivePath('/settings', '/settings')).toBe(true));
  it('child path', () => expect(isActivePath('/settings/profile', '/settings')).toBe(true));
  it('non-match', () => expect(isActivePath('/dashboard', '/settings')).toBe(false));
  it('root only matches root', () => {
    expect(isActivePath('/', '/')).toBe(true);
    expect(isActivePath('/dashboard', '/')).toBe(false);
  });
});

describe('withQueryParams', () => {
  it('adds params to clean URL', () => {
    const result = withQueryParams('/search', { q: 'hello', page: 1 });
    expect(result).toContain('q=hello');
    expect(result).toContain('page=1');
  });

  it('merges with existing params', () => {
    const result = withQueryParams('/search?q=old', { q: 'new' });
    expect(result).toBe('/search?q=new');
  });
});

describe('isAbsoluteUrl', () => {
  it('https URL', () => expect(isAbsoluteUrl('https://example.com')).toBe(true));
  it('http URL', () => expect(isAbsoluteUrl('http://example.com')).toBe(true));
  it('relative path', () => expect(isAbsoluteUrl('/dashboard')).toBe(false));
});

describe('getOrigin', () => {
  it('returns origin for valid URL', () => {
    expect(getOrigin('https://example.com/path?q=1')).toBe('https://example.com');
  });

  it('returns null for invalid URL', () => {
    expect(getOrigin('not-a-url')).toBeNull();
  });
});
