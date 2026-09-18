import { describe, expect, it, beforeEach, vi, afterEach } from 'vitest';
import { uploadAsset } from './upload';

/**
 * rtylr#952 defect 2. uploadAsset's private getToken called
 * decodeURIComponent without the try/catch its two siblings have (client.ts
 * getToken, cookies.ts), so a cookie containing malformed percent-encoding made
 * every upload throw URIError synchronously, before any request was attempted.
 *
 * document.cookie is assigned rather than redefined: the shared test setup
 * already installs a non-configurable getter/setter pair over it.
 */
const JWT = 'aaa.bbb.ccc';

beforeEach(() => {
  document.cookie = '';
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({ url: 'https://cdn.example/a.png', key: 'a.png' }),
      text: async () => '',
    }))
  );
});
afterEach(() => vi.unstubAllGlobals());

describe('uploadAsset token parsing', () => {
  it('does not throw when the cookie has malformed percent-encoding', async () => {
    // '%E0%A4%A' is an incomplete escape, so decodeURIComponent throws URIError.
    document.cookie = `token=${JWT}%E0%A4%A`;
    const file = new File(['x'], 'a.png', { type: 'image/png' });
    await expect(uploadAsset(file)).resolves.toBeTruthy();
  });

  it('still uploads with a well-formed cookie', async () => {
    document.cookie = `token=${JWT}`;
    const file = new File(['x'], 'a.png', { type: 'image/png' });
    await expect(uploadAsset(file)).resolves.toBeTruthy();
  });
});
