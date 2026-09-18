import { MemoryServiceClient } from '@voxire/voxie/hub/memory/v1/client';
import { authHeaders } from './auth';

/**
 * The Caret panel is embedded across every app and doesn't have access to any
 * one app's `safe`/`createOneKitClient` instance, so it builds its own
 * self-contained MemoryServiceClient here using the same cookie-based auth
 * header helper the rest of this package already uses.
 */
export function createMemoryClient(apiBaseUrl: string): MemoryServiceClient {
  const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl.slice(0, -1) : apiBaseUrl;
  return new MemoryServiceClient(baseUrl, {
    fetch: (input, init) =>
      fetch(input, {
        ...init,
        credentials: 'include',
        headers: { ...init?.headers, ...authHeaders() },
      }),
  });
}
