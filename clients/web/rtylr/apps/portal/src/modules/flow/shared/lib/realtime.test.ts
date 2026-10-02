import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => {
  const fetch = vi.fn();
  const statusHandlers = new Set<(s: string) => void>();
  const state = { status: 'disconnected' as string };
  return {
    fetch,
    createOneKitFetch: vi.fn(() => fetch),
    getOneKitApiBaseUrl: vi.fn(() => 'https://api.test/'),
    statusHandlers,
    state,
    getRealtimeClient: vi.fn(() => ({
      subscribe: vi.fn(),
      getStatus: () => state.status,
      subscribeStatus: (h: (s: string) => void) => {
        statusHandlers.add(h);
        h(state.status);
        return () => statusHandlers.delete(h);
      },
    })),
  };
});

vi.mock('@voxire/api', () => ({
  createOneKitFetch: mocks.createOneKitFetch,
  getOneKitApiBaseUrl: mocks.getOneKitApiBaseUrl,
  getRealtimeClient: mocks.getRealtimeClient,
}));

import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { getRealtimeTicket, useRealtimeReconnectRefetch, useRealtimeStatus } from './realtime';

describe('getRealtimeTicket', () => {
  beforeEach(() => {
    mocks.fetch.mockReset();
  });

  it('uses the authenticated OneKit fetch to exchange a REST session for a ticket', async () => {
    mocks.fetch.mockResolvedValue({
      ok: true,
      json: async () => ({ ticket: 'opaque-single-use-ticket' }),
    });

    await expect(getRealtimeTicket()).resolves.toBe('opaque-single-use-ticket');
    expect(mocks.fetch).toHaveBeenCalledWith('https://api.test/flow/ws-ticket', {
      method: 'POST',
      headers: { Accept: 'application/json' },
    });
  });

  it('returns no ticket when the exchange is unsuccessful', async () => {
    mocks.fetch.mockResolvedValue({ ok: false, json: async () => ({}) });

    await expect(getRealtimeTicket()).resolves.toBeNull();
  });
});

describe('useRealtimeStatus', () => {
  function emit(next: string) {
    mocks.state.status = next;
    act(() => {
      mocks.statusHandlers.forEach((h) => h(next));
    });
  }

  it("starts from the client's current status rather than assuming connected", () => {
    mocks.state.status = 'connecting';
    const { result } = renderHook(() => useRealtimeStatus());

    expect(result.current).toBe('connecting');
  });

  it('re-renders when the connection drops and again when it returns', () => {
    mocks.state.status = 'connected';
    const { result } = renderHook(() => useRealtimeStatus());
    expect(result.current).toBe('connected');

    emit('reconnecting');
    expect(result.current).toBe('reconnecting');

    emit('connected');
    expect(result.current).toBe('connected');
  });

  it('unsubscribes on unmount, so a closed panel does not keep updating', () => {
    mocks.state.status = 'connected';
    const { unmount } = renderHook(() => useRealtimeStatus());
    const before = mocks.statusHandlers.size;

    unmount();

    expect(mocks.statusHandlers.size).toBe(before - 1);
  });
});

describe('useRealtimeReconnectRefetch', () => {
  function mount() {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const spy = vi.spyOn(client, 'invalidateQueries');
    renderHook(() => useRealtimeReconnectRefetch(), {
      wrapper: ({ children }: { children: React.ReactNode }) =>
        React.createElement(QueryClientProvider, { client }, children),
    });
    return spy;
  }

  function emit(next: string) {
    mocks.state.status = next;
    act(() => {
      mocks.statusHandlers.forEach((h) => h(next));
    });
  }

  function keys(spy: ReturnType<typeof mount>): string[] {
    return spy.mock.calls.map((call) => JSON.stringify(call[0]?.queryKey));
  }

  it('does not invalidate on the initial connect - there is nothing to catch up on', () => {
    mocks.state.status = 'connecting';
    const spy = mount();

    emit('connected');

    expect(spy).not.toHaveBeenCalled();
  });

  it('invalidates the core query keys on a reconnect (reconnecting -> connected)', () => {
    mocks.state.status = 'connected';
    const spy = mount();

    emit('reconnecting');
    expect(spy).not.toHaveBeenCalled();

    emit('connected');

    expect(keys(spy)).toEqual(
      expect.arrayContaining([
        JSON.stringify(['tasks']),
        JSON.stringify(['my-tasks']),
        JSON.stringify(['projects']),
        JSON.stringify(['flow-notifications']),
      ])
    );
  });

  it('does not invalidate again on a status change that is not a reconnect', () => {
    mocks.state.status = 'connected';
    const spy = mount();

    emit('reconnecting');
    emit('connected');
    spy.mockClear();

    emit('reconnecting');

    expect(spy).not.toHaveBeenCalled();
  });

  it('unsubscribes on unmount', () => {
    mocks.state.status = 'connected';
    const before = mocks.statusHandlers.size;
    const { unmount } = renderHook(() => useRealtimeReconnectRefetch(), {
      wrapper: ({ children }: { children: React.ReactNode }) =>
        React.createElement(
          QueryClientProvider,
          { client: new QueryClient({ defaultOptions: { queries: { retry: false } } }) },
          children
        ),
    });

    unmount();

    expect(mocks.statusHandlers.size).toBe(before);
  });
});
