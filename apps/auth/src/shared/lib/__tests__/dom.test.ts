import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  isBrowser,
  getCurrentOrigin,
  getCurrentHostname,
  setDocumentTitle,
  getQueryParam,
} from '../dom';

beforeEach(() => {
  vi.stubGlobal('window', {
    location: {
      origin: 'https://auth.rtylr.com',
      hostname: 'auth.rtylr.com',
      search: '?redirect=dash&token=abc',
    },
    matchMedia: () => ({ matches: false }),
  });
});

describe('isBrowser', () => {
  it('returns true in a browser-like environment', () => {
    expect(isBrowser()).toBe(true);
  });
});

describe('getCurrentOrigin', () => {
  it('returns the current window origin', () => {
    expect(getCurrentOrigin()).toBe('https://auth.rtylr.com');
  });
});

describe('getCurrentHostname', () => {
  it('returns the current hostname', () => {
    expect(getCurrentHostname()).toBe('auth.rtylr.com');
  });
});

describe('setDocumentTitle', () => {
  it('sets document.title', () => {
    setDocumentTitle('Login');
    expect(document.title).toBe('Login');
  });
});

describe('getQueryParam', () => {
  it('reads a query parameter from window.location.search', () => {
    expect(getQueryParam('redirect')).toBe('dash');
  });

  it('returns null for an absent parameter', () => {
    expect(getQueryParam('missing')).toBeNull();
  });
});
