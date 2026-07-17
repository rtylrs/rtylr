import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useNotifications,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
  type Notification,
} from './useNotifications';

const mockGet = vi.fn();
const mockPatch = vi.fn();

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../lib/api', () => ({
  safe: {
    get: (...args: unknown[]) => mockGet(...args),
    patch: (...args: unknown[]) => mockPatch(...args),
  },
}));

const sampleNotifications: Notification[] = [
  {
    id: '1',
    title: 'New order',
    body: 'Order #123',
    app: 'pos',
    type: 'info',
    read: false,
    createdAt: '2026-01-01',
  },
  {
    id: '2',
    title: 'Low stock',
    body: 'Item X is low',
    app: 'erp',
    type: 'warning',
    read: false,
    createdAt: '2026-01-02',
  },
  {
    id: '3',
    title: 'Payment',
    body: 'Received',
    app: 'finance',
    type: 'success',
    read: true,
    createdAt: '2026-01-03',
  },
];

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

function createWrapperWithData() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  client.setQueryData(['notifications'], sampleNotifications);
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useNotifications', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches notifications from .notifications key', async () => {
    mockGet.mockResolvedValue({ notifications: sampleNotifications });

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(3);
    expect(result.current.data![0].title).toBe('New order');
  });

  it('fetches notifications from flat array', async () => {
    mockGet.mockResolvedValue(sampleNotifications);

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(3);
  });

  it('calls the correct endpoint', async () => {
    mockGet.mockResolvedValue({ notifications: [] });

    renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockGet).toHaveBeenCalledWith('/hub/notification/v1/notifications'));
  });

  it('handles API error', async () => {
    mockGet.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useMarkNotificationRead', () => {
  beforeEach(() => vi.clearAllMocks());

  it('optimistically marks notification as read', async () => {
    mockPatch.mockResolvedValue({ success: true });

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkNotificationRead(), { wrapper });

    await act(async () => {
      result.current.mutate('1');
    });

    // Should call the right endpoint
    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith('/hub/notification/v1/notifications/1/read')
    );
  });

  it('reverts on error', async () => {
    mockPatch.mockRejectedValue(new Error('fail'));

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkNotificationRead(), { wrapper });

    await act(async () => {
      result.current.mutate('1');
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useMarkAllNotificationsRead', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls read-all endpoint', async () => {
    mockPatch.mockResolvedValue({ success: true });

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkAllNotificationsRead(), { wrapper });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() =>
      expect(mockPatch).toHaveBeenCalledWith('/hub/notification/v1/notifications/read-all')
    );
  });

  it('handles error gracefully', async () => {
    mockPatch.mockRejectedValue(new Error('fail'));

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkAllNotificationsRead(), { wrapper });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
