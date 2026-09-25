import axios, { type AxiosInstance } from 'axios';
import { resolveApiBase, resolveAuthUrl, resolveCookieDomain } from '@voxire/config';
import { safeFetch } from './result';

const getBaseUrl = (): string => {
  // Runtime config (self-host) overrides the build-time/hosted default; hosted
  // returns the same VITE_API_URL/rtylr default it always did.
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

const isLocalDevHost = (hostname: string): boolean => {
  const host = hostname.toLowerCase();
  return host === 'localhost' || host === '127.0.0.1' || /^\d+\.\d+\.\d+\.\d+$/.test(host);
};

const getCookieDomain = (): string | undefined => {
  const hostname = window.location.hostname.toLowerCase();
  if (isLocalDevHost(hostname)) return undefined;
  // Self-host: scope to the deployment's own shared parent domain (if any).
  const runtimeDomain = resolveCookieDomain();
  if (runtimeDomain) return runtimeDomain;
  if (hostname === 'rtylr.com' || hostname.endsWith('.rtylr.com')) return '.rtylr.com';
  return undefined;
};

const clearTokenCookie = (): void => {
  const domain = getCookieDomain();
  if (domain) {
    document.cookie = `token=; Max-Age=0; path=/; Domain=${domain};`;
  }
  document.cookie = 'token=; Max-Age=0; path=/;';
};

const getAuthBaseUrl = (): string => {
  // Self-host uses the runtime auth URL; hosted keeps localhost/auth.rtylr.com.
  return isLocalDevHost(window.location.hostname) ? 'http://localhost:5176' : resolveAuthUrl();
};

const getAuthUrl = (appId: string): string => `${getAuthBaseUrl()}?redirect=${appId}`;

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
      // load, which can be before runtime config (self-host) has fetched. Hosted
      // resolves to the same value `getBaseUrl()` returned at creation time.
      config.baseURL = getBaseUrl();
      const token = getToken();
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
    (error) => {
      if (error.response?.data && typeof error.response.data === 'object') {
        error.response.data = convertKeys(error.response.data);
      }
      if (error.response?.status === 401) {
        clearTokenCookie();
        if (!isOnAuthApp()) {
          window.location.href = getAuthUrl(appId);
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

export function createOneKitFetch(appId: string): typeof fetch {
  return async (input, init) => {
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
      clearTokenCookie();
      if (!isOnAuthApp()) {
        window.location.href = getAuthUrl(appId);
      }
    }

    return response;
  };
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
