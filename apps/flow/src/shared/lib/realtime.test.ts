import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => {
  const fetch = vi.fn();
  return {
    fetch,
    createOneKitFetch: vi.fn(() => fetch),
    getOneKitApiBaseUrl: vi.fn(() => 'https://api.test/'),
    getRealtimeClient: vi.fn(() => ({ subscribe: vi.fn() })),
  };
});

vi.mock('@voxire/api', () => ({
  createOneKitFetch: mocks.createOneKitFetch,
  getOneKitApiBaseUrl: mocks.getOneKitApiBaseUrl,
  getRealtimeClient: mocks.getRealtimeClient,
}));

import { getRealtimeTicket } from './realtime';

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
