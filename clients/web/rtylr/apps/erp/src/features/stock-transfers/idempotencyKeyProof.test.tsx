// @vitest-environment happy-dom
/**
 * The TS-side equivalent of the Go tripwire
 * `services/onion/internal/platform/server/http/idempotency_pos_client_proof_test.go`,
 * for the two ERP stock-transfer routes.
 *
 * That Go test covers the POS order-payment route only. It builds a raw request
 * and hands it to the middleware, so it cannot see this client at all. These
 * tests drive the real generated client through the real `@voxire/api` transport
 * and read the headers off the `RequestInit` that reaches `fetch` - the same
 * thing the Go handler reads off the wire.
 *
 * Drop the `Idempotency-Key` header at either call site and these fail.
 */
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as React from 'react';
import { useCreateStockTransfer, useUpdateStockTransferStatus } from './useStockTransfers';

let requests: { url: string; init: RequestInit }[] = [];
let originalFetch: typeof globalThis.fetch;

const ok = () =>
  new Response(JSON.stringify({ stock_transfer: { id: '1' }, message: 'ok' }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });

beforeEach(() => {
  requests = [];
  originalFetch = globalThis.fetch;
  globalThis.fetch = vi.fn(async (input, init) => {
    requests.push({ url: String(input), init: init ?? {} });
    return ok();
  }) as unknown as typeof globalThis.fetch;
});

afterEach(() => {
  globalThis.fetch = originalFetch;
});

const headerOn = (n: number): string | null =>
  new Headers(requests[n]?.init.headers ?? {}).get('Idempotency-Key');

/**
 * A client whose mutations retry, which is the case the key has to survive:
 * react-query's default is `retry: 0`, so without this the retry path is never
 * exercised and a per-attempt key would look correct.
 */
const retryingClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: 2, retryDelay: 0 },
    },
  });

const wrapper = (client: QueryClient) =>
  function Wrapper({ children }: { children: React.ReactNode }) {
    return React.createElement(QueryClientProvider, { client }, children);
  };

const draft = () => ({
  fromWarehouseId: '1',
  toWarehouseId: '2',
  notes: 'restock front counter',
  items: [{ productId: '10', quantity: 5 }],
});

const statusChange = (id = '1') => ({ id, body: { status: 'received' as const } });

/** Fails the first `failures` attempts with a transport error, then succeeds. */
const failThenSucceed = (failures: number) => {
  let attempts = 0;
  return vi.fn(async (input: unknown, init?: RequestInit) => {
    requests.push({ url: String(input), init: init ?? {} });
    attempts += 1;
    if (attempts <= failures) throw new Error('network down');
    return ok();
  }) as unknown as typeof globalThis.fetch;
};

describe('the stock transfer create POST carries Idempotency-Key on the wire', () => {
  it('sends the header the server requires, on the route that requires it', async () => {
    const { result } = renderHook(() => useCreateStockTransfer(), {
      wrapper: wrapper(retryingClient()),
    });

    await act(() => result.current.mutateAsync(draft()));

    expect(requests).toHaveLength(1);
    expect(requests[0].url).toContain('/erp/stock-transfer/v1/stock-transfers');
    expect(requests[0].init.method).toBe('POST');
    // idempotency.go answers 428 Precondition Required for this route without it.
    expect(headerOn(0)).toEqual(expect.any(String));
  });

  it('does not leak the key into the request body', async () => {
    const { result } = renderHook(() => useCreateStockTransfer(), {
      wrapper: wrapper(retryingClient()),
    });

    await act(() => result.current.mutateAsync(draft()));

    // The header is a header. The payload is the contract's, unchanged.
    expect(String(requests[0].init.body)).not.toContain('idempotency');
  });

  it('reuses one key across every retry of the same create', async () => {
    globalThis.fetch = failThenSucceed(2);
    const { result } = renderHook(() => useCreateStockTransfer(), {
      wrapper: wrapper(retryingClient()),
    });

    await act(() => result.current.mutateAsync(draft()));

    expect(requests.map((r) => r.url.split('/v1/')[1])).toEqual([
      'stock-transfers',
      'stock-transfers',
      'stock-transfers',
    ]);
    expect(headerOn(1)).toBe(headerOn(0));
    expect(headerOn(2)).toBe(headerOn(0));
  });

  it('keys two separate drafts differently', async () => {
    const { result } = renderHook(() => useCreateStockTransfer(), {
      wrapper: wrapper(retryingClient()),
    });

    await act(() => result.current.mutateAsync(draft()));
    await act(() => result.current.mutateAsync(draft()));

    expect(requests).toHaveLength(2);
    expect(headerOn(0)).not.toBe(headerOn(1));
  });
});

describe('the stock transfer status POST carries Idempotency-Key on the wire', () => {
  it('sends the header the server requires, on the route that requires it', async () => {
    const { result } = renderHook(() => useUpdateStockTransferStatus(), {
      wrapper: wrapper(retryingClient()),
    });

    await act(() => result.current.mutateAsync(statusChange()));

    expect(requests).toHaveLength(1);
    expect(requests[0].url).toContain('/erp/stock-transfer/v1/stock-transfers/1/status');
    // idempotency.go answers 428 Precondition Required for this route without it.
    expect(headerOn(0)).toEqual(expect.any(String));
  });

  it('reuses one key across every retry of the same status change', async () => {
    globalThis.fetch = failThenSucceed(2);
    const { result } = renderHook(() => useUpdateStockTransferStatus(), {
      wrapper: wrapper(retryingClient()),
    });

    await act(() => result.current.mutateAsync(statusChange()));

    expect(requests.map((r) => r.url.split('/v1/')[1])).toEqual([
      'stock-transfers/1/status',
      'stock-transfers/1/status',
      'stock-transfers/1/status',
    ]);
    expect(headerOn(1)).toBe(headerOn(0));
    expect(headerOn(2)).toBe(headerOn(0));
  });

  it('keys the same transition on two different transfers differently', async () => {
    const { result } = renderHook(() => useUpdateStockTransferStatus(), {
      wrapper: wrapper(retryingClient()),
    });

    await act(() => result.current.mutateAsync(statusChange('1')));
    await act(() => result.current.mutateAsync(statusChange('2')));

    expect(requests).toHaveLength(2);
    expect(headerOn(0)).not.toBe(headerOn(1));
  });
});
