import axios, { type AxiosInstance } from 'axios';
import { resolveApiBase, resolveAuthUrl, resolveCookieDomain } from '@voxire/config';

const getBaseUrl = (): string => {
  // Runtime config (self-host) overrides the build-time/hosted default; hosted
  // returns the same VITE_API_URL/rtylr default it always did.
  return resolveApiBase();
};

const getToken = (): string | null => {
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  const token = match ? decodeURIComponent(match[1]) : null;
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

function snakeToCamel(str: string): string {
  return str.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
}

function convertKeys(obj: unknown): unknown {
  if (Array.isArray(obj)) {
    return obj.map(convertKeys);
  }
  if (obj !== null && typeof obj === 'object' && !(obj instanceof Date)) {
    const converted: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
      converted[snakeToCamel(key)] = convertKeys(value);
    }
    return converted;
  }
  return obj;
}

function camelToSnake(str: string): string {
  return str.replace(/[A-Z]/g, (c) => '_' + c.toLowerCase());
}

function convertKeysToSnake(obj: unknown): unknown {
  if (Array.isArray(obj)) {
    return obj.map(convertKeysToSnake);
  }
  if (obj !== null && typeof obj === 'object' && !(obj instanceof Date)) {
    const converted: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(obj as Record<string, unknown>)) {
      converted[camelToSnake(key)] = convertKeysToSnake(value);
    }
    return converted;
  }
  return obj;
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
