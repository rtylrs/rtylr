import { useEffect } from 'react';
import { getRealtimeClient, type RealtimeEvent } from '@voxire/api';

/**
 * getToken reads the auth cookie set by the auth app. The cookie is
 * marked HttpOnly in production - but we intentionally use a
 * non-HttpOnly `token` cookie for the JWT so the WebSocket handshake
 * can pass it via query string. The http-only cookie for session
 * identity remains separate.
 */
function getTokenFromCookie(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|;\s*)token=([^;]+)/);
  return match ? decodeURIComponent(match[1]) : null;
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
  getToken: getTokenFromCookie,
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
