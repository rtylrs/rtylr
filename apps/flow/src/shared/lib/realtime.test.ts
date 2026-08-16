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
import { getRealtimeTicket, useRealtimeStatus } from './realtime';

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
