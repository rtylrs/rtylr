import axios, { type AxiosInstance } from 'axios';
import {
  resolveApiBase,
  resolveAppUrl,
  resolveAuthUrl,
} from '@voxire/config';
import { safeFetch } from './result';
import { markCredential, sentCredential, shouldEndSharedSession, setSessionRefresh } from './session';

const getBaseUrl = (): string => {
  return resolveApiBase();
};

const getToken = (): string | null => {
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  let token: string | null = null;
  if (match) {
    try {
      token = decodeURIComponent(match[1]);
    } catch {
      // Malformed percent-encoding must not crash every request.
      token = match[1];
    }
  }
  if (!token || token.split('.').length !== 3) return null;
  return token;
};

const safelyGetToken = (): string | null => {
  try {
    return getToken();
  } catch {
    return null;
  }
};

const isLocalDevHost = (hostname: string): boolean => {
  const host = hostname.toLowerCase();
  return host === 'localhost' || host === '127.0.0.1' || /^\d+\.\d+\.\d+\.\d+$/.test(host);
};

const getCookieDomain = (): string | undefined => {
  const hostname = window.location.hostname.toLowerCase();
  if (isLocalDevHost(hostname)) return undefined;
  if (hostname === 'rtylr.com' || hostname.endsWith('.rtylr.com')) return '.rtylr.com';
  return undefined;
};

const clearTokenCookie = (): void => {
  const domain = getCookieDomain();
  if (domain) {
    document.cookie = `token=; Max-Age=0; path=/; Domain=${domain};`;
    document.cookie = `refresh_token=; Max-Age=0; path=/; Domain=${domain};`;
  }
  document.cookie = 'token=; Max-Age=0; path=/;';
  document.cookie = 'refresh_token=; Max-Age=0; path=/;';
};

const getRefreshToken = (): string | null => {
  const match = document.cookie.match(/(?:^|;\s*)refresh_token=([^;]+)/);
  if (!match) return null;
  try {
    return decodeURIComponent(match[1]);
  } catch {
    return match[1];
  }
};

const setAccessToken = (token: string): void => {
  const domain = getCookieDomain();
  const secure = window.location.protocol === 'https:' ? '; Secure' : '';
  const domainPart = domain ? `; Domain=${domain}` : '';
  document.cookie = `token=${encodeURIComponent(token)}; Max-Age=2592000; path=/${domainPart}; SameSite=Lax${secure}`;
};

setSessionRefresh(async (signal) => {
  if (typeof window === 'undefined' || typeof document === 'undefined') return 'unavailable';
  const refreshToken = getRefreshToken();
  if (!refreshToken) return 'rejected';
  const response = await fetch(`${getBaseUrl().replace(/\/+$/, '')}/common/auth/v1/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken }),
    signal,
  });
  if (response.status === 401) return 'rejected';
  if (!response.ok) return 'unavailable';
  const data: unknown = await response.json();
  if (
    !data ||
    typeof data !== 'object' ||
    !('token' in data) ||
    typeof data.token !== 'string' ||
    !data.token
  ) return 'unavailable';
  setAccessToken(data.token);
  return 'refreshed';
});

const getAuthBaseUrl = (): string => {
  return isLocalDevHost(window.location.hostname) ? 'http://localhost:5176' : resolveAuthUrl();
};

const getAuthUrl = (appId: string): string => `${getAuthBaseUrl()}?redirect=${appId}`;

const isMFARequiredResponse = (data: unknown): boolean => {
  if (!data || typeof data !== 'object' || !('message' in data)) return false;
  const message = (data as { message?: unknown }).message;
  return message === 'MFA enrollment required' || message === 'MFA verification required';
};

const redirectToSecuritySettings = (): void => {
  const dashboardUrl = new URL(resolveAppUrl('dash'), window.location.href);
  const dashboardPath = dashboardUrl.pathname.replace(/\/+$/, '');
  const securityPath = `${dashboardPath}/security`;
  if (
    window.location.origin === dashboardUrl.origin &&
    (window.location.pathname === securityPath ||
      window.location.pathname.startsWith(`${securityPath}/`))
  ) {
    return;
  }
  dashboardUrl.pathname = securityPath;
  dashboardUrl.search = '';
  dashboardUrl.hash = '';
  window.location.href = dashboardUrl.toString();
};

// True when the current page IS the auth app. A 401 there (e.g. a wrong
// password on the login screen, or an expired cookie during the login page's
// own /user/me pre-check) must surface as an error to the calling component -
// redirecting back to the same auth app would reload the page in a loop and
// swallow the message.
const isOnAuthApp = (): boolean => {
  try {
    const base = new URL(getAuthBaseUrl());
    return window.location.protocol === base.protocol && window.location.host === base.host;
  } catch {
    return false;
  }
};

// ── 401 classification ────────────────────────────────────────────

/**
 * 401 bodies that do not mean "your token is no longer valid".
 *
 * The bearer gate answers a bad token with a bare "Unauthorized"
 * (auth_middleware.go). These are 401s the backend raises for other reasons
 * while the token is still good: unverified email, an account awaiting
 * activation, outstanding MFA, a failed authorization lookup, or a credential
 * the caller itself submitted and got wrong (a wrong password on the login
 * form, a wrong current password, an expired Google code). Signing the user
 * out of all ten apps for any of them is the bug this list exists to prevent.
 */
const NON_SESSION_401_MESSAGES = new Set([
  'authorization unavailable',
  'email verification required',
  'account activation required',
  'mfa enrollment required',
  'mfa verification required',
  'mfa verification unavailable',
  'invalid email or password',
  'google code is invalid or expired',
  'current password is incorrect',
  'internal authority token required',
]);

/** Pull a message out of a JSON or plain-text error body. */
function unauthorizedMessage(data: unknown): string | null {
  if (typeof data === 'string') {
    const trimmed = data.trim();
    return trimmed || null;
  }
  if (!data || typeof data !== 'object') return null;
  const record = data as Record<string, unknown>;
  for (const key of ['message', 'error', 'detail']) {
    const value = record[key];
    if (typeof value === 'string' && value.trim()) return value.trim();
  }
  return null;
}

/**
 * True only for a 401 that actually rejects the bearer token.
 *
 * An unrecognised body is treated as a session rejection on purpose: the only
 * safe default for a status the client does not understand is the behaviour it
 * had before, and a genuinely dead session has to keep signing the user out.
 */
function isSessionRejection(data: unknown): boolean {
  const message = unauthorizedMessage(data);
  if (message === null) return true;
  return !NON_SESSION_401_MESSAGES.has(message.toLowerCase());
}

/** Run the shared sign-out without ever throwing into a response interceptor. */
function endSharedSession(appId: string): void {
  try {
    clearTokenCookie();
    if (!isOnAuthApp()) {
      window.location.href = getAuthUrl(appId);
    }
  } catch {
    // Sign-out is best effort. A cookie or navigation failure here must not
    // replace the real 401 the caller is about to handle.
  }
}

// ── snake_case → camelCase converter ──────────────────────────────

/**
 * True only for a plain object or array, the two shapes whose keys are ours to
 * rewrite.
 *
 * Both converters used to walk anything with typeof 'object' via
 * Object.entries. FormData, File, Blob, URLSearchParams, ArrayBuffer and Map
 * have no own enumerable properties, so each was silently replaced by {}
 * (rtylr#952). Date was the one special case someone had already hit.
 *
 * A prototype check rather than a list of instanceof arms, so it covers the
 * cases nobody has hit yet, including class instances the caller did not intend
 * to have rewritten.
 */
function isRewritableContainer(value: unknown): boolean {
  if (Array.isArray(value)) return true;
  if (value === null || typeof value !== 'object') return false;
  const proto: unknown = Object.getPrototypeOf(value);
  return proto === Object.prototype || proto === null;
}

/**
 * Matches any character after an underscore, not just a lowercase one.
 *
 * The previous /_([a-z])/g left SCREAMING_CASE untouched, so a FULL_NAME key
 * arrived as FULL_NAME and every consumer reading fullName saw undefined
 * (rtylr#952). Runs of underscores collapse, and the tail is lowercased so
 * FULL_NAME becomes fullName rather than fULLNAME.
 */
function snakeToCamel(str: string): string {
  const lower = /[A-Z]/.test(str) && !/[a-z]/.test(str) ? str.toLowerCase() : str;
  return lower.replace(/_+([a-zA-Z0-9])/g, (_, c: string) => c.toUpperCase());
}

function convertKeys(obj: unknown): unknown {
  if (Array.isArray(obj)) {
    return obj.map(convertKeys);
  }
  if (!isRewritableContainer(obj)) {
    return obj;
  }
  const converted: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
    converted[snakeToCamel(key)] = convertKeys(value);
  }
  return converted;
}

/**
 * Splits on case boundaries rather than before every capital.
 *
 * The previous /[A-Z]/g turned planTypeURL into plan_type_u_r_l. Two passes:
 * an acronym run followed by a word (URLPath -> URL_Path), then a
 * lower-or-digit to upper boundary (planType -> plan_Type). Lowercasing at the
 * end gives plan_type_url.
 *
 * Measured before changing it: 0 of the 1154 distinct keys in
 * packages/voxie/api produce an acronym run when camelCased, so this is a no-op
 * on all current traffic and a guard against the next field that would have
 * been corrupted.
 */
function camelToSnake(str: string): string {
  return str
    .replace(/([A-Z]+)([A-Z][a-z0-9])/g, '$1_$2')
    .replace(/([a-z0-9])([A-Z])/g, '$1_$2')
    .toLowerCase();
}

function convertKeysToSnake(obj: unknown): unknown {
  if (Array.isArray(obj)) {
    return obj.map(convertKeysToSnake);
  }
  if (!isRewritableContainer(obj)) {
    return obj;
  }
  const converted: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
    converted[camelToSnake(key)] = convertKeysToSnake(value);
  }
  return converted;
}

/** Convert a JSON compatibility payload to the backend's snake_case wire shape. */
export function toOneKitJson<T>(value: T): T {
  return convertKeysToSnake(value) as T;
}

/** Convert a JSON compatibility response from the backend's snake_case wire shape. */
export function fromOneKitJson<T>(value: unknown): T {
  return convertKeys(value) as T;
}

// ── Client factory ────────────────────────────────────────────────

export function createApiClient(appId: string): AxiosInstance {
  const client = axios.create({
    baseURL: getBaseUrl(),
  });

  // Auth header
  client.interceptors.request.use(
    (config) => {
      // Resolve the base per-request: the singleton `api` is created at module
      config.baseURL = getBaseUrl();
      const token = getToken();
      // Recorded so the response interceptor can tell a 401 against a real
      // credential from one against a request that never sent one. Set even
      // when there is no token, so "absent" and "not credentialed" are the
      // same answer rather than a stale leftover.
      markCredential(config, token);
      if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
      }
      // Convert request body keys from camelCase to snake_case
      if (config.data && typeof config.data === 'object' && !(config.data instanceof FormData)) {
        config.data = convertKeysToSnake(config.data);
      }
      if (config.params && typeof config.params === 'object') {
        config.params = convertKeysToSnake(config.params);
      }
      return config;
    },
    (error) => Promise.reject(error)
  );

  // Response: convert snake_case keys to camelCase + handle 401
  client.interceptors.response.use(
    (response) => {
      if (response.data && typeof response.data === 'object') {
        response.data = convertKeys(response.data);
      }
      return response;
    },
    async (error) => {
      if (error.response?.data && typeof error.response.data === 'object') {
        error.response.data = convertKeys(error.response.data);
      }
      if (error.response?.status === 401) {
        // An MFA-required 401 is a step-up prompt, not a dead session: signing
        // out would discard a valid token and strand the user mid-sign-in.
        if (isMFARequiredResponse(error.response.data)) {
          redirectToSecuritySettings();
        } else {
          const config = error.config as (
            typeof error.config & {
              headers?: Record<string, unknown>;
              __voxireSessionRetried?: boolean;
            }
          ) | undefined;
          const sentToken = sentCredential(config);
          const canRetry =
            config &&
            sentToken &&
            isSessionRejection(error.response.data) &&
            !config.__voxireSessionRetried &&
            safelyGetToken() === sentToken;
          const ended = await shouldEndSharedSession(
            sentToken,
            isSessionRejection(error.response.data),
            getToken,
            () => endSharedSession(appId)
          );
          const refreshedToken = safelyGetToken();
          if (canRetry && !ended && refreshedToken && refreshedToken !== sentToken) {
            // Axios creates a new config when replaying; carry the guard on
            // the config so mergeConfig preserves it across that copy.
            config.__voxireSessionRetried = true;
            config.headers = config.headers ?? {};
            config.headers.Authorization = `Bearer ${refreshedToken}`;
            markCredential(config, refreshedToken);
            return client.request(config);
          }
        }
      }
      return Promise.reject(error);
    }
  );

  return client;
}

/**
 * Creates the explicitly scoped fallback transport for an endpoint that has
 * no generated OneKit contract yet. Callers must keep these endpoints in a
 * small adapter module and parse every response at that boundary. Generated
 * clients remain mandatory whenever a contract exists.
 */
export function createRestRouteClient(appId: string) {
  return safeFetch(createApiClient(appId));
}

export type OneKitClientOptions = {
  fetch?: typeof fetch;
  defaultHeaders?: Record<string, string>;
};

export type OneKitClientConstructor<TClient, TOptions extends OneKitClientOptions> = new (
  baseUrl: string,
  options?: TOptions
) => TClient;

export const getOneKitApiBaseUrl = (): string => getBaseUrl();

/**
 * Read a 401 body without disturbing the response the caller still has to
 * parse. A clone that fails yields null, which classifies as a session
 * rejection - the same answer a bodiless 401 gets.
 */
async function readUnauthorizedBody(response: Response): Promise<unknown> {
  try {
    return await response.clone().json();
  } catch {
    return null;
  }
}

export function createOneKitFetch(appId: string): typeof fetch {
  const send = async (input: RequestInfo | URL, init: RequestInit | undefined, retry = false): Promise<Response> => {
    const headers = new Headers(init?.headers);
    const token = getToken();
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }

    const response = await fetch(input, {
      ...init,
      headers,
    });

    if (response.status === 401) {
      // An MFA-required 401 is a step-up prompt, not a dead session: sending the
      // user to the sign-in page would discard a perfectly valid token.
      const unauthorizedBody = await readUnauthorizedBody(response);
      if (isMFARequiredResponse(unauthorizedBody)) {
        redirectToSecuritySettings();
      } else {
        const ended = await shouldEndSharedSession(
          token,
          isSessionRejection(unauthorizedBody),
          getToken,
          () => endSharedSession(appId)
        );
        const refreshedToken = getToken();
        if (!retry && !ended && token && refreshedToken && refreshedToken !== token) {
          return send(input, init, true);
        }
      }
    }

    return response;
  };
  return (input, init) => send(input, init);
}

export function createOneKitClientOptions<TOptions extends OneKitClientOptions>(
  appId: string,
  options?: TOptions
): TOptions {
  return {
    ...(options ?? ({} as TOptions)),
    fetch: options?.fetch ?? createOneKitFetch(appId),
    defaultHeaders: {
      ...options?.defaultHeaders,
    },
  };
}

export function createOneKitClient<TClient, TOptions extends OneKitClientOptions>(
  Client: OneKitClientConstructor<TClient, TOptions>,
  appId: string,
  options?: TOptions
): TClient {
  return new Client(getOneKitApiBaseUrl(), createOneKitClientOptions(appId, options));
}

/**
 * Extracts the backend-provided message from an error thrown by a generated
 * oneKit service client. onekit's `Error` type serializes as a flat
 * `{ message, status_code }` JSON body (see `onekit/http.NewError` in onion) -
 * generated clients wrap that body as a plain string on `ApiError.body`, so it
 * has to be parsed out rather than read off `.message` (which is always the
 * generic "Request failed with status N").
 */
export function getOneKitErrorMessage(error: unknown, fallback: string): string {
  if (error && typeof error === 'object' && 'body' in error && typeof error.body === 'string') {
    try {
      const parsed: unknown = JSON.parse(error.body);
      if (parsed && typeof parsed === 'object' && 'message' in parsed) {
        const message = (parsed as { message?: unknown }).message;
        if (typeof message === 'string' && message.trim()) return message;
      }
    } catch {
      // body wasn't JSON - fall through to the generic message below.
    }
  }
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

export const api: AxiosInstance = createApiClient('dash');
