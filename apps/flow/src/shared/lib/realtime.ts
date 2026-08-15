import { useEffect, useState } from 'react';
import {
  createOneKitFetch,
  getOneKitApiBaseUrl,
  getRealtimeClient,
  type RealtimeEvent,
  type RealtimeStatus,
} from '@voxire/api';

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
  // VITE_API_URL is a full http(s) URL; swap scheme for ws(s).
  const base = (import.meta.env.VITE_API_URL as string | undefined) ?? 'https://api.voxire.com/';
  const stripped = base.replace(/\/$/, '');
  const wsScheme = stripped.startsWith('https://') ? 'wss://' : 'ws://';
  return `${wsScheme}${stripped.replace(/^https?:\/\//, '')}/flow/ws`;
}

const client = getRealtimeClient({
  url: resolveWsUrl(),
  getTicket: getRealtimeTicket,
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
