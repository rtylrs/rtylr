import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import {
  useNotifications,
  useNotificationPages,
  useMarkNotificationRead,
  useMarkAllNotificationsRead,
  type Notification,
} from './useNotifications';

const mockNotificationClient = vi.hoisted(() => ({
  listNotifications: vi.fn(),
  markNotificationRead: vi.fn(),
  markAllRead: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((v) => v),
}));

vi.mock('../../lib/clients', () => ({ notificationClient: mockNotificationClient }));

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

function createWrapperWithPages() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  client.setQueryData(['notifications'], sampleNotifications);
  client.setQueryData(['notifications', 'pages'], {
    pages: [{ notifications: sampleNotifications, nextCursor: '' }],
    pageParams: [''],
  });
  const wrapper = ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
  return { wrapper, client };
}

describe('useNotifications', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches notifications from .notifications key', async () => {
    mockNotificationClient.listNotifications.mockResolvedValue({
      notifications: sampleNotifications,
    });

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(3);
    expect(result.current.data![0].title).toBe('New order');
  });

  it('preserves the Hub notification link field', async () => {
    mockNotificationClient.listNotifications.mockResolvedValue({
      notifications: [
        {
          id: '1',
          title: 'Access request',
          body: 'Review the request',
          type: 'info',
          read: false,
          createdAt: '2026-01-01',
          link: '/team',
        },
      ],
    });

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0].link).toBe('/team');
  });

  it('fetches notifications from flat array', async () => {
    mockNotificationClient.listNotifications.mockResolvedValue(sampleNotifications);

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(3);
  });

  it('parses an omitted read flag as unread', async () => {
    // The Go contract writes `read` with `omitempty`, so an unread
    // notification arrives without the key.
    mockNotificationClient.listNotifications.mockResolvedValue({
      notifications: [
        {
          id: '1',
          title: 'New order',
          body: 'Order #123',
          app: 'pos',
          type: 'info',
          createdAt: '2026-01-01',
        },
      ],
    });

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0].read).toBe(false);
  });

  it('calls the correct endpoint', async () => {
    mockNotificationClient.listNotifications.mockResolvedValue({ notifications: [] });

    renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockNotificationClient.listNotifications).toHaveBeenCalledWith({}));
  });

  it('handles API error', async () => {
    mockNotificationClient.listNotifications.mockRejectedValue(new Error('fail'));

    const { result } = renderHook(() => useNotifications(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useNotificationPages', () => {
  beforeEach(() => vi.clearAllMocks());

  it('follows the opaque next cursor when loading another page', async () => {
    mockNotificationClient.listNotifications
      .mockResolvedValueOnce({
        notifications: sampleNotifications.slice(0, 1),
        next_cursor: 'next-page',
      })
      .mockResolvedValueOnce({ notifications: sampleNotifications.slice(1), next_cursor: '' });

    const { result } = renderHook(() => useNotificationPages(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockNotificationClient.listNotifications).toHaveBeenCalledWith({
      limit: 100,
      cursor: undefined,
    });

    await waitFor(() => expect(result.current.hasNextPage).toBe(true));

    await act(async () => {
      await result.current.fetchNextPage();
    });

    await waitFor(() => expect(result.current.data?.pages).toHaveLength(2));
    expect(mockNotificationClient.listNotifications).toHaveBeenLastCalledWith({
      limit: 100,
      cursor: 'next-page',
    });
  });

  it('parses an omitted read flag as unread', async () => {
    mockNotificationClient.listNotifications.mockResolvedValue({
      notifications: [
        {
          id: '1',
          title: 'New order',
          body: 'Order #123',
          app: 'pos',
          type: 'info',
          createdAt: '2026-01-01',
        },
      ],
    });

    const { result } = renderHook(() => useNotificationPages(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.pages[0].notifications[0].read).toBe(false);
  });
});

describe('useMarkNotificationRead', () => {
  beforeEach(() => vi.clearAllMocks());

  it('optimistically marks notification as read', async () => {
    mockNotificationClient.markNotificationRead.mockResolvedValue({ success: true });

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkNotificationRead(), { wrapper });

    await act(async () => {
      result.current.mutate('1');
    });

    // Should call the right endpoint
    await waitFor(() =>
      expect(mockNotificationClient.markNotificationRead).toHaveBeenCalledWith({ id: '1' })
    );
  });

  it('reverts on error', async () => {
    mockNotificationClient.markNotificationRead.mockRejectedValue(new Error('fail'));

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkNotificationRead(), { wrapper });

    await act(async () => {
      result.current.mutate('1');
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('restores the paged cache on error', async () => {
    mockNotificationClient.markNotificationRead.mockRejectedValue(new Error('fail'));
    mockNotificationClient.listNotifications.mockResolvedValue({
      notifications: sampleNotifications,
    });

    const { wrapper, client } = createWrapperWithPages();
    const { result } = renderHook(() => useMarkNotificationRead(), { wrapper });

    await act(async () => {
      result.current.mutate('1');
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const pages = client.getQueryData<{ pages: { notifications: Notification[] }[] }>([
      'notifications',
      'pages',
    ]);
    expect(pages?.pages[0]?.notifications.find((n) => n.id === '1')?.read).toBe(false);
  });
});

describe('useMarkAllNotificationsRead', () => {
  beforeEach(() => vi.clearAllMocks());

  it('calls read-all endpoint', async () => {
    mockNotificationClient.markAllRead.mockResolvedValue({ success: true });

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkAllNotificationsRead(), { wrapper });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() => expect(mockNotificationClient.markAllRead).toHaveBeenCalledWith({}));
  });

  it('handles error gracefully', async () => {
    mockNotificationClient.markAllRead.mockRejectedValue(new Error('fail'));

    const wrapper = createWrapperWithData();
    const { result } = renderHook(() => useMarkAllNotificationsRead(), { wrapper });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it('restores the paged cache on error', async () => {
    mockNotificationClient.markAllRead.mockRejectedValue(new Error('fail'));
    mockNotificationClient.listNotifications.mockResolvedValue({
      notifications: sampleNotifications,
    });

    const { wrapper, client } = createWrapperWithPages();
    const { result } = renderHook(() => useMarkAllNotificationsRead(), { wrapper });

    await act(async () => {
      result.current.mutate();
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    const pages = client.getQueryData<{ pages: { notifications: Notification[] }[] }>([
      'notifications',
      'pages',
    ]);
    expect(pages?.pages[0]?.notifications.filter((n) => !n.read)).toHaveLength(2);
  });
});
