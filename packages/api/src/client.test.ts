import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createApiClient } from './client';

// We need to mock import.meta.env
vi.stubGlobal('import', { meta: { env: { VITE_API_URL: 'https://api.test.com/' } } });

beforeEach(() => {
  vi.restoreAllMocks();
});

function mockCookie(value: string) {
  return vi.spyOn(document, 'cookie', 'get').mockReturnValue(value);
}

describe('createApiClient', () => {
  it('returns an axios instance', () => {
    const client = createApiClient('hub');
    expect(client).toBeDefined();
    expect(typeof client.get).toBe('function');
    expect(typeof client.post).toBe('function');
    expect(typeof client.put).toBe('function');
    expect(typeof client.delete).toBe('function');
  });

  it('client has baseURL set', () => {
    const client = createApiClient('hub');
    expect(client.defaults.baseURL).toBeDefined();
    expect(typeof client.defaults.baseURL).toBe('string');
  });

  it('request interceptor adds Authorization header when token exists', async () => {
    mockCookie('token=my-test-token');
    const client = createApiClient('hub');

    // Access the request interceptor through the config
    const config = {
      headers: {} as any,
    };

    // Get interceptors - they're stored in axios
    const interceptors = (client.interceptors.request as any).handlers;
    expect(interceptors.length).toBeGreaterThan(0);

    // Execute the fulfilled handler
    const result = interceptors[0].fulfilled(config);
    expect(result.headers['Authorization']).toBe('Bearer my-test-token');
  });

  it('request interceptor does not add header when no token', async () => {
    mockCookie('');
    const client = createApiClient('hub');

    const config = {
      headers: {} as any,
    };

    const interceptors = (client.interceptors.request as any).handlers;
    const result = interceptors[0].fulfilled(config);
    expect(result.headers['Authorization']).toBeUndefined();
  });

  it('response interceptor handles 401 by redirecting', async () => {
    const cookieSetter = vi.spyOn(document, 'cookie', 'set');
    const client = createApiClient('hub');

    const interceptors = (client.interceptors.response as any).handlers;
    expect(interceptors.length).toBeGreaterThan(0);

    // The rejected handler handles 401s
    const error = { response: { status: 401 } };
    try {
      await interceptors[0].rejected(error);
    } catch {
      // Expected to reject
    }

    expect(cookieSetter).toHaveBeenCalledWith('token=; Max-Age=0; path=/;');
  });

  it('response interceptor passes through non-401 errors', async () => {
    const client = createApiClient('hub');

    const interceptors = (client.interceptors.response as any).handlers;
    const error = { response: { status: 500 } };

    await expect(interceptors[0].rejected(error)).rejects.toEqual(error);
  });

  it('response interceptor passes through successful responses', () => {
    const client = createApiClient('hub');

    const interceptors = (client.interceptors.response as any).handlers;
    const response = { status: 200, data: { ok: true } };

    const result = interceptors[0].fulfilled(response);
    expect(result).toEqual(response);
  });

  it('creates different clients for different appIds', () => {
    const hubClient = createApiClient('hub');
    const authClient = createApiClient('auth');
    expect(hubClient).not.toBe(authClient);
  });

  it('client has interceptors registered', () => {
    const client = createApiClient('test-app');
    const reqHandlers = (client.interceptors.request as any).handlers;
    const resHandlers = (client.interceptors.response as any).handlers;
    expect(reqHandlers.length).toBe(1);
    expect(resHandlers.length).toBe(1);
  });

  it('handles token with special characters in cookie', () => {
    mockCookie('token=eyJhbGciOiJIUzI1NiJ9.test.sig');
    const client = createApiClient('hub');

    const config = { headers: {} as any };
    const interceptors = (client.interceptors.request as any).handlers;
    const result = interceptors[0].fulfilled(config);
    expect(result.headers['Authorization']).toBe('Bearer eyJhbGciOiJIUzI1NiJ9.test.sig');
  });
});
