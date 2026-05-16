import { describe, expect, it } from 'vitest';
import { joinUrl, trimTrailingSlash } from './url';

describe('trimTrailingSlash', () => {
  it('drops a single trailing slash', () => {
    expect(trimTrailingSlash('https://x.com/api/')).toBe('https://x.com/api');
  });
  it('leaves a bare root slash alone', () => {
    expect(trimTrailingSlash('/')).toBe('/');
  });
});

describe('joinUrl', () => {
  it('dedups the boundary slash', () => {
    expect(joinUrl('https://x.com/', '/y')).toBe('https://x.com/y');
  });
  it('inserts a missing boundary slash', () => {
    expect(joinUrl('https://x.com', 'y')).toBe('https://x.com/y');
  });
});
