import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  safeParseUrl,
  getSearchParam,
  appendSearchParam,
  isSafeRedirect,
  stripTrailingSlash,
} from '../url';

beforeEach(() => {
  vi.stubGlobal('window', { location: { origin: 'https://auth.rtylr.com' } });
});

describe('safeParseUrl', () => {
  it('parses a valid URL', () => {
    expect(safeParseUrl('https://example.com/path')?.pathname).toBe('/path');
  });

  it('returns null for an invalid URL', () => {
    expect(safeParseUrl('not a url')).toBeNull();
  });
});

describe('getSearchParam', () => {
  it('extracts a param from a full URL', () => {
    expect(getSearchParam('https://example.com?redirect=pos', 'redirect')).toBe('pos');
  });

  it('extracts a param from a bare query string', () => {
    expect(getSearchParam('redirect=dash&token=abc', 'redirect')).toBe('dash');
  });

  it('returns null when param is absent', () => {
    expect(getSearchParam('https://example.com?foo=bar', 'redirect')).toBeNull();
  });
});

describe('appendSearchParam', () => {
  it('appends a new param to a URL', () => {
    expect(appendSearchParam('https://example.com/', 'redirect', 'pos')).toBe(
      'https://example.com/?redirect=pos'
    );
  });

  it('returns the original string for an invalid URL', () => {
    expect(appendSearchParam('not-a-url', 'key', 'val')).toBe('not-a-url');
  });
});

describe('isSafeRedirect', () => {
  it('accepts same-origin URLs', () => {
    expect(isSafeRedirect('https://auth.rtylr.com/callback')).toBe(true);
  });

  it('accepts an explicitly allowed origin', () => {
    expect(isSafeRedirect('https://dash.rtylr.com/', ['https://dash.rtylr.com'])).toBe(true);
  });

  it('rejects unknown external URLs', () => {
    expect(isSafeRedirect('https://evil.com/steal')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isSafeRedirect('')).toBe(false);
  });
});

describe('stripTrailingSlash', () => {
  it('removes a trailing slash', () => {
    expect(stripTrailingSlash('/docs/')).toBe('/docs');
  });

  it('does not remove the slash from the root path', () => {
    expect(stripTrailingSlash('/')).toBe('/');
  });
});
