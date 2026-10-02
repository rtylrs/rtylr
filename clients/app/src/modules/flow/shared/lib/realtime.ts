import { useEffect, useRef, useState } from 'react';
import { useQueryClient, type QueryKey } from '@tanstack/react-query';
import {
  createOneKitFetch,
  getOneKitApiBaseUrl,
  getRealtimeClient,
  type RealtimeDiagnosticEvent,
  type RealtimeEvent,
  type RealtimeStatus,
} from '@voxire/api';
import { Sentry } from '@voxire/observability';

const oneKitFetch = createOneKitFetch('flow');

/** Exchange the authenticated REST session for a one-use WS ticket. */
export async function getRealtimeTicket(): Promise<string | null> {
  const baseUrl = getOneKitApiBaseUrl().replace(/\/$/, '');
  try {
    const response = await oneKitFetch(`${baseUrl}/flow/ws-ticket`, {
      method: 'POST',
      headers: { Accept: 'application/json' },
    });
    if (!response.ok) return null;

    const body: unknown = await response.json();
    if (!body || typeof body !== 'object' || !('ticket' in body)) return null;
    const ticket = (body as { ticket?: unknown }).ticket;
    return typeof ticket === 'string' && ticket.length > 0 ? ticket : null;
  } catch {
    return null;
  }
}

function resolveWsUrl(): string {
  const url = new URL(
    `${getOneKitApiBaseUrl().replace(/\/$/, '')}/flow/ws`,
    window.location.origin
  );
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
  return url.toString();
}

/** Report non-fatal connection diagnostics without treating them as crashes. */
function reportRealtimeDiagnostic(event: RealtimeDiagnosticEvent): void {
  Sentry.addBreadcrumb({
    category: 'realtime',
    level: 'warning',
    message:
      event.type === 'malformed_message'
        ? 'Dropped a malformed realtime message'
        : `Realtime reconnect failed ${event.attempt} times in a row`,
    data: event.type === 'repeated_reconnect_failures' ? { attempt: event.attempt } : undefined,
  });
}

const client = getRealtimeClient({
  url: resolveWsUrl(),
  getTicket: getRealtimeTicket,
  onDiagnostic: reportRealtimeDiagnostic,
});

/**
 * Subscribe to one or more realtime topics for the lifetime of the
 * component. The handler fires on every matching event. Subscriptions
 * are reference-counted - if ten components subscribe to
 * `project:5`, the server only sees one subscription.
 *
 * The handler should be stable across renders (wrap in useCallback or
 * keep logic inline) to avoid rapid subscribe/unsubscribe churn.
 */
export function useRealtime(topics: string[], handler: (event: RealtimeEvent) => void): void {
  useEffect(() => {
    if (topics.length === 0) return;
    const unsubs = topics.map((t) => client.subscribe(t, handler));
    return () => {
      unsubs.forEach((u) => u());
    };
    // We intentionally depend on `topics.join` so callers can pass a
    // fresh array literal without thrashing subscriptions.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [topics.join('|'), handler]);
}

/**
 * Observe the realtime connection state.
 *
 * Surfaces that would otherwise show stale data silently use this to say so.
 * `reconnecting` means the client is retrying with backoff and writes still
 * work; `disconnected` means it is not coming back on its own.
 */
export function useRealtimeStatus(): RealtimeStatus {
  const [status, setStatus] = useState<RealtimeStatus>(() => client.getStatus());
  useEffect(() => client.subscribeStatus(setStatus), []);
  return status;
}

/**
 * Query keys broad enough to catch up on state that changed while the
 * socket was down. Events carry no payload, so a
 * reconnect can't replay what was missed - the shared client re-subscribes
 * to the same topics on `ws.onopen` but never learns what fired while
 * disconnected. Rather than trying to track exactly what was missed, we
 * invalidate the lists every realtime consumer already keys off of.
 */
const RECONNECT_INVALIDATE_KEYS: readonly QueryKey[] = [
  ['tasks'],
  ['my-tasks'],
  ['projects'],
  ['flow-notifications'],
];

/**
 * Invalidate the app's core query keys on *reconnect* - a transition back to
 * `connected` from `reconnecting` - not on the initial connect, which has
 * nothing to catch up on.
 *
 * Events missed while disconnected cannot be replayed.
 *
 * Mount once near the app root - `FlowLayout` does this alongside
 * `useLayoutRealtime`.
 */
export function useRealtimeReconnectRefetch(): void {
  const qc = useQueryClient();
  const previousStatus = useRef<RealtimeStatus>(client.getStatus());

  useEffect(
    () =>
      client.subscribeStatus((status) => {
        if (status === 'connected' && previousStatus.current === 'reconnecting') {
          for (const queryKey of RECONNECT_INVALIDATE_KEYS) {
            qc.invalidateQueries({ queryKey });
          }
        }
        previousStatus.current = status;
      }),
    [qc]
  );
}
