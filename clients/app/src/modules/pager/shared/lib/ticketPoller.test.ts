import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { TICKET_POLL_INTERVAL_MS, createTicketPoller, isStaleAnswer } from './ticketPoller';

interface Answer {
  status: string;
  serverTime: string;
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

function setup(fetchTicket: () => Promise<Answer>) {
  const onTicket = vi.fn();
  const onNotFound = vi.fn();
  const onError = vi.fn();
  const poller = createTicketPoller<Answer>({ fetchTicket, onTicket, onNotFound, onError });
  return { poller, onTicket, onNotFound, onError };
}

function setVisibility(state: DocumentVisibilityState): void {
  Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => state });
  document.dispatchEvent(new Event('visibilitychange'));
}

const waiting = (serverTime: string): Answer => ({ status: 'waiting', serverTime });
const ready = (serverTime: string): Answer => ({ status: 'ready', serverTime });

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
  setVisibility('visible');
});

describe('isStaleAnswer', () => {
  it.each([
    ['2026-11-19T10:00:05.000Z', '2026-11-19T10:00:04.999Z', true],
    ['2026-11-19T10:00:05.000Z', '2026-11-19T10:00:05.000Z', false],
    ['2026-11-19T10:00:05.000Z', '2026-11-19T10:00:05.001Z', false],
    [undefined, '2026-11-19T10:00:05.000Z', false],
    ['2026-11-19T10:00:05.000Z', undefined, false],
    ['2026-11-19T10:00:05.000Z', 'not a time', false],
  ])('shown %s, incoming %s: stale=%s', (shown, incoming, expected) => {
    expect(isStaleAnswer(shown, incoming)).toBe(expected);
  });
});

describe('createTicketPoller', () => {
  it('asks straight away and then every three seconds', async () => {
    const fetchTicket = vi.fn(async () => waiting('2026-11-19T10:00:00.000Z'));
    const { poller, onTicket } = setup(fetchTicket);

    poller.start();
    await vi.advanceTimersByTimeAsync(0);
    expect(fetchTicket).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS * 2);
    expect(fetchTicket).toHaveBeenCalledTimes(3);
    expect(onTicket).toHaveBeenCalledTimes(3);

    poller.stop();
  });

  it('drops an answer older than the one already shown', async () => {
    const slow = deferred<Answer>();
    const fetchTicket = vi
      .fn<() => Promise<Answer>>()
      .mockReturnValueOnce(slow.promise)
      .mockResolvedValueOnce(ready('2026-11-19T10:00:06.000Z'));
    const { poller, onTicket } = setup(fetchTicket);

    poller.start();
    poller.pollNow();
    await vi.advanceTimersByTimeAsync(0);
    expect(onTicket).toHaveBeenLastCalledWith(ready('2026-11-19T10:00:06.000Z'));

    // The first request lands late, carrying what was true before the order was ready.
    slow.resolve(waiting('2026-11-19T10:00:03.000Z'));
    await vi.advanceTimersByTimeAsync(0);

    expect(onTicket).toHaveBeenCalledTimes(1);
    poller.stop();
  });

  it('does not stack requests while one is still in the air', async () => {
    const fetchTicket = vi.fn(() => new Promise<Answer>(() => {}));
    const { poller } = setup(fetchTicket);

    poller.start();
    await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS * 3);

    expect(fetchTicket).toHaveBeenCalledTimes(1);
    poller.stop();
  });

  it('asks at once when the page becomes visible again', async () => {
    const fetchTicket = vi.fn(async () => waiting('2026-11-19T10:00:00.000Z'));
    const { poller } = setup(fetchTicket);
    poller.start();
    await vi.advanceTimersByTimeAsync(0);

    setVisibility('hidden');
    expect(fetchTicket).toHaveBeenCalledTimes(1);

    setVisibility('visible');
    expect(fetchTicket).toHaveBeenCalledTimes(2);
    poller.stop();
  });

  it.each(['collected', 'expired'])('stops once the ticket is %s', async (status) => {
    const fetchTicket = vi.fn(async () => ({ status, serverTime: '2026-11-19T10:00:00.000Z' }));
    const { poller, onTicket } = setup(fetchTicket);

    poller.start();
    await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS * 3);
    setVisibility('visible');

    expect(fetchTicket).toHaveBeenCalledTimes(1);
    expect(onTicket).toHaveBeenCalledTimes(1);
  });

  it('stops and reports a ticket that no longer exists', async () => {
    const fetchTicket = vi.fn(async () => {
      throw { statusCode: 404, code: 'ticket_not_found' };
    });
    const { poller, onNotFound, onError } = setup(fetchTicket);

    poller.start();
    await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS * 3);

    expect(fetchTicket).toHaveBeenCalledTimes(1);
    expect(onNotFound).toHaveBeenCalledTimes(1);
    expect(onError).not.toHaveBeenCalled();
  });

  it('keeps polling through a failed request', async () => {
    const fetchTicket = vi
      .fn<() => Promise<Answer>>()
      .mockRejectedValueOnce({ statusCode: 503 })
      .mockResolvedValue(ready('2026-11-19T10:00:06.000Z'));
    const { poller, onTicket, onError } = setup(fetchTicket);

    poller.start();
    await vi.advanceTimersByTimeAsync(0);
    expect(onError).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS);
    expect(onTicket).toHaveBeenCalledWith(ready('2026-11-19T10:00:06.000Z'));
    poller.stop();
  });

  it('ignores an answer that arrives after it was stopped', async () => {
    const pending = deferred<Answer>();
    const { poller, onTicket } = setup(() => pending.promise);

    poller.start();
    poller.stop();
    pending.resolve(ready('2026-11-19T10:00:06.000Z'));
    await vi.advanceTimersByTimeAsync(0);

    expect(onTicket).not.toHaveBeenCalled();
  });
});
