import { isGone } from './apiError';

export const TICKET_POLL_INTERVAL_MS = 3_000;

const CLOSED_STATUSES = new Set(['collected', 'expired']);

export function isOpenStatus(status: string | undefined): boolean {
  return status === 'waiting' || status === 'ready';
}

/**
 * True when `incoming` was built by the server before the answer already on
 * screen. Answers can overtake each other on a phone's network, and showing
 * the older one would put a ready order back to waiting.
 */
export function isStaleAnswer(shown: string | undefined, incoming: string | undefined): boolean {
  if (!shown || !incoming) return false;
  const shownAt = Date.parse(shown);
  const incomingAt = Date.parse(incoming);
  if (Number.isNaN(shownAt) || Number.isNaN(incomingAt)) return false;
  return incomingAt < shownAt;
}

export interface TicketPollerOptions<T> {
  fetchTicket: () => Promise<T>;
  onTicket: (ticket: T) => void;
  /** The ticket is gone; polling has stopped. */
  onNotFound: () => void;
  /** A poll failed; polling carries on. */
  onError: (error: unknown) => void;
  intervalMs?: number;
}

export interface TicketPoller {
  start: () => void;
  stop: () => void;
  /** Polls now, even with a request already in the air. */
  pollNow: () => void;
}

export function createTicketPoller<T extends { status?: string; serverTime?: string }>(
  options: TicketPollerOptions<T>
): TicketPoller {
  const intervalMs = options.intervalMs ?? TICKET_POLL_INTERVAL_MS;
  let running = false;
  let inFlight = 0;
  let timer: ReturnType<typeof setInterval> | undefined;
  let shownServerTime: string | undefined;

  function stop(): void {
    running = false;
    clearInterval(timer);
    document.removeEventListener('visibilitychange', onVisibilityChange);
  }

  async function poll(force: boolean): Promise<void> {
    if (!running || (inFlight > 0 && !force)) return;
    inFlight += 1;
    try {
      const ticket = await options.fetchTicket();
      if (!running || isStaleAnswer(shownServerTime, ticket.serverTime)) return;
      shownServerTime = ticket.serverTime ?? shownServerTime;
      options.onTicket(ticket);
      if (CLOSED_STATUSES.has(ticket.status ?? '')) stop();
    } catch (error) {
      if (!running) return;
      // A 404 without the pager's code is a route a deploy has not served
      // yet, not a missing ticket: keep polling.
      if (isGone(error)) {
        stop();
        options.onNotFound();
      } else {
        options.onError(error);
      }
    } finally {
      inFlight -= 1;
    }
  }

  // A phone freezes a background tab's timers, so the first thing to do when
  // the customer looks at the page again is ask.
  function onVisibilityChange(): void {
    if (document.visibilityState === 'visible') void poll(true);
  }

  return {
    start: () => {
      if (running) return;
      running = true;
      document.addEventListener('visibilitychange', onVisibilityChange);
      timer = setInterval(() => void poll(false), intervalMs);
      void poll(false);
    },
    stop,
    pollNow: () => void poll(true),
  };
}
