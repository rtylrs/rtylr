import { describe, expect, it, beforeEach, vi, afterEach } from 'vitest';
import { signAssetUrl, uploadAsset } from './upload';

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

describe('signAssetUrl', () => {
  it('asks the CDN to sign the file with the user token and returns the link', async () => {
    document.cookie = `token=${JWT}`;
    const fetchMock = vi.fn(async () => ({
      ok: true,
      status: 200,
      json: async () => ({ url: 'https://cdn.example/files/x?sig=s', expires_at: 99 }),
      text: async () => '',
    }));
    vi.stubGlobal('fetch', fetchMock);

    await expect(
      signAssetUrl('https://cdn.example/files/x', { cdnBaseUrl: 'https://cdn.example/' })
    ).resolves.toEqual({ url: 'https://cdn.example/files/x?sig=s', expiresAt: 99 });

    const [endpoint, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(endpoint).toBe('https://cdn.example/sign');
    expect((init.headers as Record<string, string>).Authorization).toBe(`Bearer ${JWT}`);
    expect(JSON.parse(init.body as string)).toEqual({
      url: 'https://cdn.example/files/x',
      expires_in: 3600,
    });
  });

  it('reports a refused signature as an error', async () => {
    document.cookie = `token=${JWT}`;
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => ({ ok: false, status: 404, text: async () => 'file not found' }))
    );

    await expect(signAssetUrl('https://cdn.example/files/x')).rejects.toThrow('file not found');
  });
});
