import { describe, expect, it } from 'vitest';
import { isHttpUrl } from './isUrl';

describe('isHttpUrl', () => {
  it('accepts http(s) URLs', () => {
    expect(isHttpUrl('http://x.com')).toBe(true);
    expect(isHttpUrl('https://x.com')).toBe(true);
  });
  it('rejects other schemes', () => {
    expect(isHttpUrl('mailto:hi@x.com')).toBe(false);
    expect(isHttpUrl('file:///tmp')).toBe(false);
  });
  it('rejects junk', () => {
    expect(isHttpUrl('not a url')).toBe(false);
  });
});
