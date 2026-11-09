import type { QueryClient } from '@tanstack/react-query';
import type { RealtimeEvent } from '@voxire/api';

/** How long after this tab changes something its own events count as echoes. */
const ECHO_WINDOW_MS = 5000;

/**
 * Whether a realtime event is this tab's own change coming back. The mutation
 * that made it already refreshed this tab, and refetching again flickers an
 * optimistic update that is still in flight. Skipping every event the user
 * caused went too far: their other tabs and devices made no change themselves
 * and never refreshed, so only a tab that changed something just now skips.
 */
export function isOwnEcho(
  qc: QueryClient,
  e: RealtimeEvent,
  currentUserId: string | null | undefined,
  now: number = Date.now()
): boolean {
  if (!e.actor_id || !currentUserId || String(e.actor_id) !== currentUserId) return false;
  return qc
    .getMutationCache()
    .getAll()
    .some((m) => m.state.status === 'pending' || now - m.state.submittedAt < ECHO_WINDOW_MS);
}
