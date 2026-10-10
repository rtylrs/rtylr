import { CaretRuntimeServiceClient } from '@voxire/voxie/hub/caret/v1/client';
import { authHeaders } from './auth';

/** A self-contained Caret runtime client, built like the memory client. */
export function createRuntimeClient(apiBaseUrl: string): CaretRuntimeServiceClient {
  const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl.slice(0, -1) : apiBaseUrl;
  return new CaretRuntimeServiceClient(baseUrl, {
    fetch: (input, init) =>
      fetch(input, {
        ...init,
        credentials: 'include',
        headers: { ...init?.headers, ...authHeaders() },
      }),
  });
}
