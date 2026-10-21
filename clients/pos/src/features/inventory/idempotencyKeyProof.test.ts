/**
 * The TS-side equivalent of the Go tripwire
 * `services/onion/internal/platform/server/http/idempotency_pos_client_proof_test.go`,
 * for the stock transfer route.
 *
 * The Go test builds a raw request and hands it to the middleware, so it only
 * sees the headers a caller chose to put on that request; it cannot see this
 * app's client at all. This test drives the real `recordStockAction` the
 * inventory screen itself calls, through the real generated client and the real
 * `@voxire/api` transport, and reads the header off the `RequestInit` that
 * reaches `fetch` - the same thing the Go handler reads off the wire.
 *
 * Drop the `Idempotency-Key` header in stockActionClient.ts and this fails.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { buildStockAction, submitStockAction, type StockActionSubmission } from './inventoryWorkflow';
import { recordStockAction, sendStockAction } from './stockActionClient';

let requests: { url: string; init: RequestInit }[] = [];
let originalFetch: typeof globalThis.fetch;

beforeEach(() => {
  requests = [];
  originalFetch = globalThis.fetch;
  globalThis.fetch = vi.fn(async (input, init) => {
    requests.push({ url: String(input), init: init ?? {} });
    // The generated clients decode the backend's snake_case wire shape, not the
    // camelCase objects the app sees.
    return new Response(JSON.stringify({ message: 'Transfer recorded' }), {
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

const transfer = (itemId: number, reason = 'restock front counter') =>
  buildStockAction({
    itemId,
    action: 'transfer',
    quantityInput: '5',
    reason,
    branchId: 1,
    toBranchId: 2,
    employeeId: 7,
  });

describe('the POS stock transfer POST carries Idempotency-Key on the wire', () => {
  it('sends the header the server requires, on the route that requires it', async () => {
    await recordStockAction(transfer(10));

    expect(requests).toHaveLength(1);
    expect(requests[0].url).toContain('/pos/stock/v1/transfer');
    expect(requests[0].init.method).toBe('POST');
    // idempotency.go answers 428 Precondition Required for this route without it.
    expect(headerOn(0)).toEqual(expect.any(String));
  });

  it('sends the key minted for the action, unchanged, to the wire', async () => {
    const seen: string[] = [];
    await submitStockAction(transfer(10), (submission: StockActionSubmission) => {
      seen.push(submission.idempotencyKey);
      return sendStockAction(submission);
    });

    expect(headerOn(0)).toBe(seen[0]);
  });

  it('keys two separate transfers differently', async () => {
    await recordStockAction(transfer(10));
    await recordStockAction(transfer(11));

    expect(requests).toHaveLength(2);
    expect(headerOn(0)).toEqual(expect.any(String));
    expect(headerOn(1)).toEqual(expect.any(String));
    expect(headerOn(0)).not.toBe(headerOn(1));
  });

  it('reuses one key for every attempt the action makes', async () => {
    // A flaky transport retrying the same action re-sends the same submission,
    // so the wire sees one key and the server can dedup the replay.
    const sent: StockActionSubmission[] = [];
    await submitStockAction(transfer(10), (submission) => {
      sent.push(submission);
      return sendStockAction(submission);
    });
    await sendStockAction(sent[0]);
    await sendStockAction(sent[0]);

    expect(requests).toHaveLength(3);
    expect(headerOn(0)).toBe(headerOn(1));
    expect(headerOn(1)).toBe(headerOn(2));
  });
});
