import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { createApiClient, createOneKitFetch } from './client';
import {
  resetSessionInvalidationForTests,
  setSessionRefresh,
  type SessionRefreshResult,
} from './session';

vi.stubGlobal('import', {
  meta: { env: { VITE_API_URL: 'https://api.test.com/' } },
});

const TOKEN_A = 'eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.sig_a';
const TOKEN_B = 'eyJhbGciOiJub25lIn0.eyJ1c2VyX2lkIjoiMSJ9.sig_b';

function setTokenCookie(token: string): void {
  document.cookie = `token=${token}; path=/`;
}

function readTokenCookie(): string | null {
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

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

/**
 * Counts writes that CLEAR the token cookie, so "the shared cookie is cleared
 * exactly once" is observable. happy-dom deletes the cookie synchronously, so
 * reading it back can only prove the clear happened, never how many times.
 * Setting a session in a test is therefore not counted.
 *
 * The proxy replaces the global binding because happy-dom defines `cookie` as a
 * non-configurable own property on the document instance.
 */
function trackCookieWrites(): { count: () => number; stop: () => void; makeThrow: () => void } {
  const real = globalThis.document;
  let clears = 0;
  let throwOnWrite = false;
  const proxy = new Proxy(real, {
    get(target, prop, receiver) {
      if (throwOnWrite && prop === 'cookie') {
        throw new Error('cookie jar unavailable');
      }
      return Reflect.get(target, prop, receiver);
    },
    set(target, prop, value, receiver) {
      if (prop === 'cookie') {
        if (throwOnWrite) throw new Error('cookie jar unavailable');
        if (String(value).startsWith('token=;')) clears++;
      }
      return Reflect.set(target, prop, value, receiver);
    },
  });
  globalThis.document = proxy;
  return {
    count: () => clears,
    stop: () => {
      globalThis.document = real;
    },
    makeThrow: () => {
      throwOnWrite = true;
    },
  };
}

type RequestConfig = { headers: Record<string, string> };

/** The credential the request interceptor recorded, as axios stored it. */
function runRequestInterceptor(
  client: ReturnType<typeof createApiClient>,
  config: RequestConfig = { headers: {} }
): RequestConfig {
  const handlers = (
    client.interceptors.request as unknown as {
      handlers: { fulfilled: (c: RequestConfig) => RequestConfig }[];
    }
  ).handlers;
  return handlers[0]!.fulfilled(config);
}

function runResponseRejection(client: ReturnType<typeof createApiClient>, error: object) {
  const handlers = (
    client.interceptors.response as unknown as { handlers: { rejected: (e: object) => Promise<never> }[] }
  ).handlers;
  return handlers[0]!.rejected(error);
}

/** A 401 as axios delivers it, carrying the config the interceptor mutated. */
function rejected401(config: RequestConfig, data?: unknown) {
  return { config, response: { status: 401, data: data ?? { message: 'Unauthorized' } } };
}

let cookieTracker: ReturnType<typeof trackCookieWrites>;

beforeEach(() => {
  clearAllCookies();
  resetSessionInvalidationForTests();
  vi.restoreAllMocks();
  cookieTracker = trackCookieWrites();
});

afterEach(() => {
  cookieTracker.stop();
  resetSessionInvalidationForTests();
});

describe('401 handling: requests that were not credentialed', () => {
  it('does not sign out when a 401 comes back from a request that sent no token', async () => {
    const client = createApiClient('crm');
    const config = runRequestInterceptor(client);
    expect(config.headers).toEqual({});

    await runResponseRejection(client, rejected401(config)).catch(() => undefined);

    expect(cookieTracker.count()).toBe(0);
  });

  it('keeps a session established after an uncredentialed request was issued', async () => {
    const client = createApiClient('recruit');
    // The public careers endpoint is fetched before sign-in finishes, so the
    // request went out with no Authorization header.
    const config = runRequestInterceptor(client);
    expect(config.headers).toEqual({});

    // The user signs in elsewhere while that request is still in flight.
    setTokenCookie(TOKEN_A);

    await runResponseRejection(client, rejected401(config)).catch(() => undefined);

    expect(readTokenCookie()).toBe(TOKEN_A);
    expect(cookieTracker.count()).toBe(0);
  });
});

describe('401 handling: 401s that are not about the session', () => {
  it.each([
    'Email verification required',
    'Account activation required',
    'MFA enrollment required',
    'MFA verification required',
    'MFA verification unavailable',
    'Authorization unavailable',
    'Invalid email or password',
    'Current password is incorrect',
    'Google code is invalid or expired',
  ])('keeps the shared session for "%s"', async (message) => {
    setTokenCookie(TOKEN_A);
    const client = createApiClient('dash');
    const config = runRequestInterceptor(client);

    await runResponseRejection(client, rejected401(config, { message })).catch(() => undefined);

    expect(readTokenCookie()).toBe(TOKEN_A);
    expect(cookieTracker.count()).toBe(0);
  });

  it('keeps the shared session when a wrong password is submitted on the auth app', async () => {
    setTokenCookie(TOKEN_A);
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValue(
        new Response(JSON.stringify({ message: 'Invalid email or password' }), { status: 401 })
      );

    await createOneKitFetch('auth')('https://api.test.com/common/auth/v1/login', { method: 'POST' });

    expect(fetchSpy).toHaveBeenCalled();
    expect(readTokenCookie()).toBe(TOKEN_A);
  });

  it('still returns the response when its 401 body cannot be read', async () => {
    setTokenCookie(TOKEN_A);
    // An unparseable body leaves the message unknown, so this signs out - the
    // pre-existing default - but the caller must still receive its response.
    const response = new Response('<html>gateway</html>', { status: 401 });
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(response);

    const received = await createOneKitFetch('dash')('https://api.test.com/hub/user/v1/users/me');

    expect(received).toBe(response);
    await vi.waitFor(() => expect(readTokenCookie()).toBeNull());
  });
});

describe('401 handling: a genuine session rejection', () => {
  it('refreshes an expired access token and retries a OneKit request once', async () => {
    setTokenCookie(TOKEN_A);
    setSessionRefresh(async () => {
      setTokenCookie(TOKEN_B);
      return 'refreshed';
    });
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(JSON.stringify({ message: 'Unauthorized' }), { status: 401 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));

    const response = await createOneKitFetch('dash')('https://api.test.com/hub/user/v1/users/me');

    expect(response.status).toBe(200);
    expect(fetchSpy).toHaveBeenCalledTimes(2);
    const retryHeaders = new Headers(fetchSpy.mock.calls[1]?.[1]?.headers);
    expect(retryHeaders.get('Authorization')).toBe(`Bearer ${TOKEN_B}`);
    expect(readTokenCookie()).toBe(TOKEN_B);
    expect(cookieTracker.count()).toBe(0);
  });

  it('signs out once and redirects to auth', async () => {
    setTokenCookie(TOKEN_A);
    const client = createApiClient('erp');
    const config = runRequestInterceptor(client);

    await runResponseRejection(client, rejected401(config)).catch(() => undefined);
    await vi.waitFor(() => expect(window.location.href).toContain('localhost:5176'));

    expect(readTokenCookie()).toBeNull();
  });

  it('clears the shared cookie exactly once for N concurrent 401s', async () => {
    setTokenCookie(TOKEN_A);
    const client = createApiClient('hr');
    const configs = Array.from({ length: 6 }, () =>
      runRequestInterceptor(client)
    );

    await Promise.all(
      configs.map((config) => runResponseRejection(client, rejected401(config)).catch(() => undefined))
    );
    await vi.waitFor(() => expect(window.location.href).toContain('localhost:5176'));

    expect(readTokenCookie()).toBeNull();
    // One clear, not six. On localhost there is no parent domain to clear
    // separately, so a single sign-out is exactly one cookie write.
    expect(cookieTracker.count()).toBe(1);
  });

  it('re-arms after a new sign-in so a later expiry can sign out again', async () => {
    setTokenCookie(TOKEN_A);
    const client = createApiClient('hr');
    await runResponseRejection(
      client,
      rejected401(runRequestInterceptor(client))
    ).catch(() => undefined);
    await vi.waitFor(() => expect(readTokenCookie()).toBeNull());

    setTokenCookie(TOKEN_B);
    const next = runRequestInterceptor(client);
    await runResponseRejection(client, rejected401(next)).catch(() => undefined);
    await vi.waitFor(() => expect(readTokenCookie()).toBeNull());

    // Two genuine sign-outs in total, one per session.
    expect(cookieTracker.count()).toBe(2);
  });
});

describe('401 handling: stale responses and refresh races', () => {
  it('does not sign out when the rejected token has already been replaced', async () => {
    setTokenCookie(TOKEN_A);
    const client = createApiClient('flow');
    const stale = runRequestInterceptor(client);

    // A sign-in in another tab rewrote the cookie before the 401 came back.
    setTokenCookie(TOKEN_B);

    await runResponseRejection(client, rejected401(stale)).catch(() => undefined);

    expect(readTokenCookie()).toBe(TOKEN_B);
    expect(cookieTracker.count()).toBe(0);
  });

  it('does not sign out on 401s while a session refresh is in flight', async () => {
    setTokenCookie(TOKEN_A);
    let release: (result: SessionRefreshResult) => void = () => {};
    setSessionRefresh(() => new Promise<SessionRefreshResult>((resolve) => (release = resolve)));

    const client = createApiClient('dash');
    const configs = Array.from({ length: 3 }, () =>
      runRequestInterceptor(client)
    );

    const pending = Promise.all(
      configs.map((config) => runResponseRejection(client, rejected401(config)).catch(() => undefined))
    );
    // Give the in-flight checks a chance to run before they resolve.
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(readTokenCookie()).toBe(TOKEN_A);
    expect(cookieTracker.count()).toBe(0);

    // The refresh replaced the cookie; the 401s belonged to the old token.
    setTokenCookie(TOKEN_B);
    release('refreshed');
    await pending;

    expect(readTokenCookie()).toBe(TOKEN_B);
    expect(cookieTracker.count()).toBe(0);
  });

  it('signs out exactly once when the refresh definitively fails', async () => {
    setTokenCookie(TOKEN_A);
    let calls = 0;
    setSessionRefresh(() => {
      calls++;
      return Promise.resolve('rejected');
    });

    const client = createApiClient('dash');
    const configs = Array.from({ length: 4 }, () =>
      runRequestInterceptor(client)
    );

    await Promise.all(
      configs.map((config) => runResponseRejection(client, rejected401(config)).catch(() => undefined))
    );
    await vi.waitFor(() => expect(window.location.href).toContain('localhost:5176'));

    // Single-flighted: four 401s, one refresh attempt, one sign-out.
    expect(calls).toBe(1);
    expect(readTokenCookie()).toBeNull();
    expect(cookieTracker.count()).toBe(1);
  });

  it('keeps the session through a temporary refresh failure and retries refresh later', async () => {
    setTokenCookie(TOKEN_A);
    let calls = 0;
    setSessionRefresh(() => {
      calls++;
      if (calls === 1) return Promise.reject(new Error('network down'));
      setTokenCookie(TOKEN_B);
      return Promise.resolve('refreshed');
    });

    const client = createApiClient('dash');
    await runResponseRejection(
      client,
      rejected401(runRequestInterceptor(client))
    ).catch(() => undefined);
    expect(readTokenCookie()).toBe(TOKEN_A);
    expect(cookieTracker.count()).toBe(0);

    await runResponseRejection(
      client,
      rejected401(runRequestInterceptor(client))
    ).catch(() => undefined);
    expect(readTokenCookie()).toBe(TOKEN_B);

    expect(calls).toBe(2);
    expect(cookieTracker.count()).toBe(0);
  });

  it('a failed sign-out does not reject the caller 401', async () => {
    setTokenCookie(TOKEN_A);
    const client = createApiClient('dash');
    const error = rejected401(runRequestInterceptor(client));

    cookieTracker.makeThrow();

    await expect(runResponseRejection(client, error)).rejects.toBe(error);
  });
});
