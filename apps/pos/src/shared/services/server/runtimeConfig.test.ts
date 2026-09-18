import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { fetchRuntimeConfig } from './runtimeConfig';

describe('fetchRuntimeConfig', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  it('fetches runtime-config from normalized API base URL and returns auth/upload URLs', async () => {
    const mockResponse = {
      urls: {
        auth: 'https://auth.my-server.local',
        upload: 'https://upload.my-server.local',
      },
    };

    (globalThis.fetch as Mock).mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    });

    const result = await fetchRuntimeConfig('http://192.168.1.100:8000');
    expect(globalThis.fetch).toHaveBeenCalledWith('http://192.168.1.100:8000/runtime-config', {
      method: 'GET',
      headers: { Accept: 'application/json' },
      signal: expect.any(AbortSignal),
    });

    expect(result).toEqual({
      authUrl: 'https://auth.my-server.local',
      uploadUrl: 'https://upload.my-server.local',
    });
  });

  it('returns nulls if the urls section is empty or missing from response', async () => {
    (globalThis.fetch as Mock).mockResolvedValue({
      ok: true,
      json: async () => ({}),
    });

    const result = await fetchRuntimeConfig('http://192.168.1.100:8000/');
    expect(result).toEqual({
      authUrl: null,
      uploadUrl: null,
    });
  });

  it('returns nulls and does not throw if the response status is not ok', async () => {
    (globalThis.fetch as Mock).mockResolvedValue({
      ok: false,
      status: 500,
    });

    const result = await fetchRuntimeConfig('http://192.168.1.100:8000/');
    expect(result).toEqual({
      authUrl: null,
      uploadUrl: null,
    });
  });

  it('returns nulls and does not throw if the fetch call rejects/throws an exception', async () => {
    (globalThis.fetch as Mock).mockRejectedValue(new Error('Network offline'));

    const result = await fetchRuntimeConfig('http://192.168.1.100:8000/');
    expect(result).toEqual({
      authUrl: null,
      uploadUrl: null,
    });
  });
});
