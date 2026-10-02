// @vitest-environment happy-dom
/**
 * The TS-side equivalent of the Go tripwire
 * `services/onion/internal/platform/server/http/idempotency_pos_client_proof_test.go`,
 * for the two finance routes the predicate also covers.
 *
 * That Go test covers the POS order-payment route only. It builds a raw request
 * and hands it to the middleware, so it cannot see these clients at all. These
 * tests drive the real generated clients through the real `@voxire/api`
 * transport and read the headers off the `RequestInit` that reaches `fetch` -
 * the same thing the Go handler reads off the wire.
 *
 * Drop the `Idempotency-Key` header at either call site and these fail.
 */
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import * as React from 'react';
import { useRecordPayment } from './useInvoices';
import { useSendInvoice } from '../../shared/hooks/queries/useInvoices';

let requests: { url: string; init: RequestInit }[] = [];
let originalFetch: typeof globalThis.fetch;

beforeEach(() => {
  requests = [];
  originalFetch = globalThis.fetch;
  globalThis.fetch = vi.fn(async (input, init) => {
    requests.push({ url: String(input), init: init ?? {} });
    return new Response(JSON.stringify({ message: 'ok' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
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

const payment = (id = '42') => ({
  id,
  amount: 250,
  currency: 'USD',
  paymentDate: '2026-06-01',
  method: 'card',
  reference: 'TXN-42',
});

/** Fails the first `failures` attempts with a transport error, then succeeds. */
const failThenSucceed = (failures: number) => {
  let attempts = 0;
  return vi.fn(async (input, init) => {
    requests.push({ url: String(input), init: (init as RequestInit) ?? {} });
    attempts += 1;
    if (attempts <= failures) throw new Error('network down');
    return new Response(JSON.stringify({ message: 'ok' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  }) as unknown as typeof globalThis.fetch;
};

describe('the invoice payment POST carries Idempotency-Key on the wire', () => {
  it('sends the header the server requires, on the route that requires it', async () => {
    const { result } = renderHook(() => useRecordPayment(), { wrapper: wrapper(retryingClient()) });

    await act(() => result.current.mutateAsync(payment()));

    expect(requests).toHaveLength(1);
    expect(requests[0].url).toContain('/finance/invoices/payments/v1/invoices/42/payment');
    expect(requests[0].init.method).toBe('POST');
    // idempotency.go answers 428 Precondition Required for this route without it.
    expect(headerOn(0)).toEqual(expect.any(String));
  });

  it('reuses one key across every retry of the same payment', async () => {
    globalThis.fetch = failThenSucceed(2);
    const { result } = renderHook(() => useRecordPayment(), { wrapper: wrapper(retryingClient()) });

    await act(() => result.current.mutateAsync(payment()));

    // Three attempts, one key: a retry must not mint a second key, or the
    // server cannot tell it from a second, real payment.
    expect(requests).toHaveLength(3);
    expect(headerOn(0)).toEqual(expect.any(String));
    expect(headerOn(1)).toBe(headerOn(0));
    expect(headerOn(2)).toBe(headerOn(0));
  });

  it('keys two separate payments differently', async () => {
    const { result } = renderHook(() => useRecordPayment(), { wrapper: wrapper(retryingClient()) });

    await act(() => result.current.mutateAsync(payment('42')));
    await act(() => result.current.mutateAsync(payment('43')));

    expect(requests).toHaveLength(2);
    expect(headerOn(0)).toEqual(expect.any(String));
    expect(headerOn(1)).toEqual(expect.any(String));
    expect(headerOn(0)).not.toBe(headerOn(1));
  });
});

describe('the invoice send POST carries Idempotency-Key on the wire', () => {
  it('sends the header the server requires, on the route that requires it', async () => {
    const { result } = renderHook(() => useSendInvoice(), { wrapper: wrapper(retryingClient()) });

    await act(() => result.current.mutateAsync({ invoiceId: '9' }));

    expect(requests).toHaveLength(1);
    expect(requests[0].url).toContain('/finance/invoices/v1/invoices/9/send');
    expect(requests[0].init.method).toBe('POST');
    // idempotency.go answers 428 Precondition Required for this route without it.
    expect(headerOn(0)).toEqual(expect.any(String));
  });

  it('reuses one key across every retry of the same send', async () => {
    globalThis.fetch = failThenSucceed(2);
    const { result } = renderHook(() => useSendInvoice(), { wrapper: wrapper(retryingClient()) });

    await act(() => result.current.mutateAsync({ invoiceId: '9' }));

    expect(requests).toHaveLength(3);
    expect(headerOn(1)).toBe(headerOn(0));
    expect(headerOn(2)).toBe(headerOn(0));
  });

  it('keys two separate sends differently', async () => {
    const { result } = renderHook(() => useSendInvoice(), { wrapper: wrapper(retryingClient()) });

    await act(() => result.current.mutateAsync({ invoiceId: '9' }));
    await act(() => result.current.mutateAsync({ invoiceId: '9' }));

    expect(requests).toHaveLength(2);
    expect(headerOn(0)).not.toBe(headerOn(1));
  });
});
