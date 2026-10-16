import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useWebhooks, useCreateWebhook, useDeleteWebhook } from './useWebhooks';

const mockDeveloperClient = vi.hoisted(() => ({
  listWebhooks: vi.fn(),
  createWebhook: vi.fn(),
  deleteWebhook: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
  getOneKitErrorMessage: vi.fn((_err, fallback) => fallback),
}));

vi.mock('../../lib/clients', () => ({ developerClient: mockDeveloperClient }));

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

  it('fetches the contract webhook fields without inventing display-only fields', async () => {
    mockDeveloperClient.listWebhooks.mockResolvedValue({
      webhooks: [
        {
          id: 1,
          url: 'https://example.com/wh',
          events: ['order_created'],
          isActive: false,
          createdAt: '2026-09-01T12:00:00Z',
          lastTriggeredAt: '2026-09-23T11:00:00Z',
        },
      ],
    });

    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useWebhooks(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]).toMatchObject({
      url: 'https://example.com/wh',
      isActive: false,
      lastTriggeredAt: '2026-09-23T11:00:00Z',
    });
    expect(result.current.data?.[0]).not.toHaveProperty('status');
    expect(result.current.data?.[0]).not.toHaveProperty('secretPreview');
  });

  it('creates a webhook successfully and trims the destination', async () => {
    mockDeveloperClient.createWebhook.mockResolvedValue({ id: 1, secret: 'sec_123' });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateWebhook(), { wrapper });

    act(() => {
      result.current.mutate({ url: '  https://example.com/wh  ', events: ['order_created'] });
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockDeveloperClient.createWebhook).toHaveBeenCalledWith({
      url: 'https://example.com/wh',
      events: ['order_created'],
    });
    expect(result.current.data).toEqual({ id: '1', secret: 'sec_123' });
  });

  it('rejects event names that the live dispatcher cannot publish', async () => {
    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateWebhook(), { wrapper });

    await expect(
      result.current.mutateAsync({ url: 'https://example.com/wh', events: ['order.created'] })
    ).rejects.toThrow('Unsupported webhook event');
    expect(mockDeveloperClient.createWebhook).not.toHaveBeenCalled();
  });

  it('rejects a non-HTTP destination before calling the API', async () => {
    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useCreateWebhook(), { wrapper });

    await expect(
      result.current.mutateAsync({ url: 'not-a-url', events: ['order_created'] })
    ).rejects.toThrow('valid HTTP(S) URL');
    expect(mockDeveloperClient.createWebhook).not.toHaveBeenCalled();
  });

  it('deletes a webhook successfully', async () => {
    mockDeveloperClient.deleteWebhook.mockResolvedValue({ success: true });

    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });
    const wrapper = createWrapper(client);
    const { result } = renderHook(() => useDeleteWebhook(), { wrapper });

    act(() => {
      result.current.mutate('12');
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockDeveloperClient.deleteWebhook).toHaveBeenCalledWith({ id: '12' });
  });
});
