import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useApiKeys, useCreateApiKey, useRevokeApiKey } from './useApiKeys';

const mockDeveloperClient = vi.hoisted(() => ({
  listApiKeys: vi.fn(),
  createApiKey: vi.fn(),
  revokeApiKey: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../lib/clients', () => ({ developerClient: mockDeveloperClient }));

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
    mockDeveloperClient.listApiKeys.mockResolvedValue({
      keys: [{ id: 1, name: 'Test Key', token: 'key_123' }],
    });

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
    mockDeveloperClient.createApiKey.mockResolvedValue({
      id: 1,
      key: 'key_abc',
      message: 'Created',
    });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateApiKey(), { wrapper });

    act(() => {
      result.current.mutate({ name: 'New Key', scopes: ['read'] });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockDeveloperClient.createApiKey).toHaveBeenCalledWith({
      name: 'New Key',
      scopes: ['read'],
    });
  });

  it('revokes an api key successfully', async () => {
    mockDeveloperClient.revokeApiKey.mockResolvedValue({ message: 'Revoked' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useRevokeApiKey(), { wrapper });

    act(() => {
      result.current.mutate('12');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockDeveloperClient.revokeApiKey).toHaveBeenCalledWith({ id: '12' });
  });
});
