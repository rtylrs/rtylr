import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  createApiClient,
  createOneKitClient,
  createOneKitFetch,
  getOneKitErrorMessage,
} from './client';

// We need to mock import.meta.env
vi.stubGlobal('import', {
  meta: { env: { VITE_API_URL: 'https://api.test.com/' } },
});

function clearAllCookies(): void {
  const names = new Set(
    document.cookie
      .split(';')
      .map((c) => c.split('=')[0]?.trim())
      .filter((n): n is string => Boolean(n))
  );
  names.add('token');
  for (const name of names) {
    document.cookie = `${name}=; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; path=/`;
  }
}

beforeEach(() => {
  clearAllCookies();
  vi.restoreAllMocks();
});

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
    document.cookie = 'token=eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.fake_sig';
    const client = createApiClient('hub');

    const config = {
      headers: {} as Record<string, string>,
    };

    const interceptors = (client.interceptors.request as any).handlers;
    expect(interceptors.length).toBeGreaterThan(0);

    const result = interceptors[0].fulfilled(config);
    expect(result.headers['Authorization']).toBe(
      'Bearer eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.fake_sig'
    );
  });

  it('request interceptor does not add header when no token', async () => {
    const client = createApiClient('hub');

    const config = {
      headers: {} as Record<string, string>,
    };

    const interceptors = (client.interceptors.request as any).handlers;
    const result = interceptors[0].fulfilled(config);
    expect(result.headers['Authorization']).toBeUndefined();
  });

  it('request interceptor converts body and query params to backend JSON names', () => {
    const client = createApiClient('hub');

    const config = {
      headers: {} as Record<string, string>,
      data: {
        clientName: 'Voxire',
        lineItems: [{ unitPrice: 12 }],
      },
      params: {
        projectId: 'flow-1',
        dateFrom: '2026-06-01',
      },
    };

    const interceptors = (client.interceptors.request as any).handlers;
    const result = interceptors[0].fulfilled(config);

    expect(result.data).toEqual({
      client_name: 'Voxire',
      line_items: [{ unit_price: 12 }],
    });
    expect(result.params).toEqual({
      project_id: 'flow-1',
      date_from: '2026-06-01',
    });
  });

  it('response interceptor handles 401 by redirecting to auth', async () => {
    document.cookie = 'token=eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.fake_sig';
    const client = createApiClient('hub');

    const interceptors = (client.interceptors.response as any).handlers;
    expect(interceptors.length).toBeGreaterThan(0);

    // Run the request interceptor first so the config carries the credential
    // this 401 is a response to. A hand-built error with no config describes a
    // request that never sent a token, which is no longer a sign-out signal.
    const reqInterceptors = (client.interceptors.request as any).handlers;
    const config = reqInterceptors[0].fulfilled({ headers: {} });

    const error = { config, response: { status: 401 } };
    try {
      await interceptors[0].rejected(error);
    } catch {
      // Expected to reject
    }

    // After 401, the token cookie should be expired (Max-Age=0 removes it)
    // and window.location.href should point to auth
    await vi.waitFor(() => expect(window.location.href).toContain('localhost:5176'));
  });

  it('response interceptor passes through non-401 errors', async () => {
    const client = createApiClient('hub');

    const interceptors = (client.interceptors.response as any).handlers;
    const error = { response: { status: 500 } };

    await expect(interceptors[0].rejected(error)).rejects.toEqual(error);
  });

  it('response interceptor passes through successful responses', () => {
    const client = createApiClient('test-app');

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
    document.cookie = 'token=eyJhbGciOiJIUzI1NiJ9.test.sig';
    const client = createApiClient('hub');

    const config = { headers: {} as Record<string, string> };
    const interceptors = (client.interceptors.request as any).handlers;
    const result = interceptors[0].fulfilled(config);
    expect(result.headers['Authorization']).toBe('Bearer eyJhbGciOiJIUzI1NiJ9.test.sig');
  });
});

describe('createOneKitFetch', () => {
  it('adds Authorization header when token exists', async () => {
    document.cookie = 'token=eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.fake_sig';
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{}'));

    await createOneKitFetch('hub')('https://api.test.com/demo', {
      headers: {
        'X-Test': 'yes',
      },
    });

    const [, init] = fetchSpy.mock.calls[0]!;
    const headers = new Headers(init?.headers);
    expect(headers.get('Authorization')).toBe(
      'Bearer eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.fake_sig'
    );
    expect(headers.get('X-Test')).toBe('yes');
  });

  it('redirects to auth on 401', async () => {
    document.cookie = 'token=eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.fake_sig';
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{}', { status: 401 }));

    await createOneKitFetch('hub')('https://api.test.com/demo');

    expect(window.location.href).toContain('localhost:5176');
  });
});

describe('createOneKitClient', () => {
  it('creates generated clients with base URL and oneKit options', () => {
    class DemoClient {
      constructor(
        public baseUrl: string,
        public options?: { fetch?: typeof fetch; defaultHeaders?: Record<string, string> }
      ) {}
    }

    const client = createOneKitClient(DemoClient, 'hub', {
      defaultHeaders: { 'X-Demo': '1' },
    });

    expect(client.baseUrl).toBeDefined();
    expect(client.options?.fetch).toBeDefined();
    expect(client.options?.defaultHeaders).toEqual({ 'X-Demo': '1' });
  });
});

describe('getOneKitErrorMessage', () => {
  it('parses the message out of a oneKit ApiError JSON body', () => {
    const error = {
      message: 'Request failed with status 401',
      body: '{"message":"Invalid email or password","status_code":401}',
    };
    expect(getOneKitErrorMessage(error, 'fallback')).toBe('Invalid email or password');
  });

  it('falls back to error.message when the body is not JSON', () => {
    const error = new Error('network down');
    expect(getOneKitErrorMessage(error, 'fallback')).toBe('network down');
  });

  it('falls back to the provided default for unrecognized errors', () => {
    expect(getOneKitErrorMessage('nope', 'fallback')).toBe('fallback');
  });
});
