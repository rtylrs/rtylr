import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useApiKeys, useCreateApiKey, useRevokeApiKey } from './useApiKeys';

const mockGet = vi.fn();
const mockPost = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
  },
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper(client: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useApiKeys', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches api keys successfully', async () => {
    mockGet.mockResolvedValue({ keys: [{ id: 1, name: 'Test Key', token: 'key_123' }] });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useApiKeys(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0].name).toBe('Test Key');
  });

  it('creates an api key successfully', async () => {
    mockPost.mockResolvedValue({ id: 1, key: 'key_abc', message: 'Created' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateApiKey(), { wrapper });

    act(() => {
      result.current.mutate({ name: 'New Key', scopes: ['read'] });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockPost).toHaveBeenCalledWith('/hub/developer/v1/api-keys', {
      name: 'New Key',
      scopes: ['read'],
    });
  });

  it('revokes an api key successfully', async () => {
    mockPost.mockResolvedValue({ message: 'Revoked' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useRevokeApiKey(), { wrapper });

    act(() => {
      result.current.mutate('12');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockPost).toHaveBeenCalledWith('/hub/developer/v1/api-keys/revoke', { id: 12 });
  });
});
