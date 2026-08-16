import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useWebhooks, useCreateWebhook, useDeleteWebhook } from './useWebhooks';

const mockGet = vi.fn();
const mockPost = vi.fn();
const mockDelete = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
  },
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper(client: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useWebhooks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches webhooks successfully', async () => {
    mockGet.mockResolvedValue({
      webhooks: [{ id: 1, url: 'https://example.com/wh', events: ['order.created'] }],
    });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useWebhooks(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0].url).toBe('https://example.com/wh');
  });

  it('creates a webhook successfully', async () => {
    mockPost.mockResolvedValue({ id: 1, secret: 'sec_123' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateWebhook(), { wrapper });

    act(() => {
      result.current.mutate({ url: 'https://example.com/wh', events: ['order.created'] });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockPost).toHaveBeenCalledWith('/hub/developer/v1/webhooks', {
      url: 'https://example.com/wh',
      events: ['order.created'],
    });
  });

  it('deletes a webhook successfully', async () => {
    mockDelete.mockResolvedValue({ success: true });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useDeleteWebhook(), { wrapper });

    act(() => {
      result.current.mutate('12');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockDelete).toHaveBeenCalledWith('/hub/developer/v1/webhooks/12');
  });
});
