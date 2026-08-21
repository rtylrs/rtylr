import { ok, err, Result, ResultAsync } from 'neverthrow';
import type { AxiosInstance, AxiosError, AxiosRequestConfig } from 'axios';

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

export interface ApiError {
  /** HTTP status code (0 for network errors) */
  status: number;
  /** Machine-readable error code from the API, if any */
  code: string;
  /** Human-readable message safe to show in UI */
  message: string;
  /** Original error for logging / debugging */
  cause?: unknown;
}

/** Narrow union for common failure modes */
export type ApiErrorCode =
  | 'NETWORK_ERROR'
  | 'TIMEOUT'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'VALIDATION'
  | 'CONFLICT'
  | 'RATE_LIMITED'
  | 'SERVER_ERROR'
  | 'UNKNOWN';

// ---------------------------------------------------------------------------
// Error parsing
// ---------------------------------------------------------------------------

function resolveCode(status: number): ApiErrorCode {
  if (status === 0) return 'NETWORK_ERROR';
  if (status === 401) return 'UNAUTHORIZED';
  if (status === 403) return 'FORBIDDEN';
  if (status === 404) return 'NOT_FOUND';
  if (status === 409) return 'CONFLICT';
  if (status === 422) return 'VALIDATION';
  if (status === 429) return 'RATE_LIMITED';
  if (status >= 500) return 'SERVER_ERROR';
  return 'UNKNOWN';
}

function fallbackMessage(status: number): string {
  if (status === 401) return 'Session expired. Please sign in again.';
  if (status === 403) return 'You do not have permission to perform this action.';
  if (status === 404) return 'The requested resource was not found.';
  if (status === 422) return 'Please check your input and try again.';
  if (status === 429) return 'Too many requests. Please wait a moment.';
  if (status >= 500) return 'Something went wrong on our end. Please try again.';
  return 'An unexpected error occurred.';
}

function messageFrom(value: unknown): string | null {
  if (typeof value === 'string') {
    const trimmed = value.trim();
    return trimmed || null;
  }
  if (!value || typeof value !== 'object') return null;

  const record = value as Record<string, unknown>;
  return (
    messageFrom(record.message) ?? messageFrom(record.error) ?? messageFrom(record.detail) ?? null
  );
}

function parseAxiosError(error: AxiosError<any>): ApiError {
  if (error.code === 'ECONNABORTED' || error.code === 'ETIMEDOUT') {
    return {
      status: 0,
      code: 'TIMEOUT',
      message: 'Request timed out. Please try again.',
      cause: error,
    };
  }

  if (!error.response) {
    return {
      status: 0,
      code: 'NETWORK_ERROR',
      message: 'Unable to reach the server. Check your connection.',
      cause: error,
    };
  }

  const status = error.response.status;
  const data = error.response.data;
  const message = messageFrom(data) ?? fallbackMessage(status);

  return { status, code: resolveCode(status), message, cause: error };
}

function toApiError(error: unknown): ApiError {
  if (error && typeof error === 'object' && 'isAxiosError' in error) {
    return parseAxiosError(error as AxiosError<any>);
  }
  return {
    status: 0,
    code: 'UNKNOWN',
    message: messageFrom(error) ?? 'An unexpected error occurred.',
    cause: error,
  };
}

// ---------------------------------------------------------------------------
// Safe API client wrapper
// ---------------------------------------------------------------------------

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export interface SafeApi {
  get<T = any>(url: string, config?: AxiosRequestConfig): ResultAsync<T, ApiError>;
  post<T = any>(url: string, data?: unknown, config?: AxiosRequestConfig): ResultAsync<T, ApiError>;
  put<T = any>(url: string, data?: unknown, config?: AxiosRequestConfig): ResultAsync<T, ApiError>;
  patch<T = any>(
    url: string,
    data?: unknown,
    config?: AxiosRequestConfig
  ): ResultAsync<T, ApiError>;
  delete<T = any>(url: string, config?: AxiosRequestConfig): ResultAsync<T, ApiError>;
}

/**
 * Wraps an Axios instance so every call returns `ResultAsync<T, ApiError>`
 * instead of throwing.
 *
 * Usage:
 * ```ts
 * import { createApiClient } from '@voxire/api';
 * import { safeFetch } from '@voxire/result';
 *
 * const api = createApiClient('crm');
 * const safe = safeFetch(api);
 *
 * const result = await safe.get<Customer[]>('/crm/customers');
 * if (result.isOk()) {
 *   console.log(result.value); // Product[]
 * } else {
 *   toast.error(result.error.message);
 * }
 * ```
 */
export function safeFetch(client: AxiosInstance): SafeApi {
  return {
    get: <T>(url: string, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(
        client.get<T>(url, config).then((r) => r.data),
        toApiError
      ),

    post: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(
        client.post<T>(url, data, config).then((r) => r.data),
        toApiError
      ),

    put: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(
        client.put<T>(url, data, config).then((r) => r.data),
        toApiError
      ),

    patch: <T>(url: string, data?: unknown, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(
        client.patch<T>(url, data, config).then((r) => r.data),
        toApiError
      ),

    delete: <T>(url: string, config?: AxiosRequestConfig) =>
      ResultAsync.fromPromise(
        client.delete<T>(url, config).then((r) => r.data),
        toApiError
      ),
  };
}

/**
 * Adapts a generated OneKit Promise to the ResultAsync shape used by the
 * existing React Query consumers. The transport remains the generated client;
 * this helper only keeps error handling consistent while callers migrate.
 */
export function oneKitResult<T>(operation: () => Promise<T>): ResultAsync<T, ApiError> {
  return ResultAsync.fromPromise(Promise.resolve().then(operation), toApiError);
}

// ---------------------------------------------------------------------------
// React Query integration helper
// ---------------------------------------------------------------------------

/**
 * Unwraps a ResultAsync for use inside React Query's queryFn / mutationFn.
 * Returns the success value or throws ApiError (which React Query catches
 * and exposes via `error` on the query/mutation object).
 *
 * Usage:
 * ```ts
 * useQuery({
 *   queryKey: ['products'],
 *   queryFn: () => unwrap(safe.get<Product[]>('/erp/product/v1/products')),
 * });
 * ```
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export async function unwrap<T = any>(result: ResultAsync<T, ApiError>): Promise<T> {
  let r: Result<T, ApiError>;
  try {
    r = await result;
  } catch (e: unknown) {
    // ResultAsync should always resolve to Ok | Err, but if the internal
    // promise rejects (e.g. the error-mapper itself threw), normalise the
    // rejection into a typed ApiError so TanStack Query transitions to its
    // error state instead of hanging in isPending forever.
    throw toApiError(e);
  }
  if (r.isOk()) return r.value;
  throw r.error;
}

// Re-export neverthrow primitives so consumers only need @voxire/api
export { ok, err, Result, ResultAsync };
export type { AxiosInstance };
