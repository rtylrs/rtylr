/**
 * Tiny analytics shim for Caret panel interactions. Currently a no-op in
 * production — wired up so the rest of the codebase can call
 * `logEvent(...)` and we can hook PostHog/Segment later without an audit.
 */

export interface CaretEvent {
  name: string;
  props?: Record<string, unknown>;
}

type Listener = (event: CaretEvent) => void;

const listeners = new Set<Listener>();

export function logEvent(name: string, props?: Record<string, unknown>): void {
  const event: CaretEvent = { name, props };
  for (const fn of listeners) {
    try {
      fn(event);
    } catch {
      // ignore — single listener failure shouldn't poison the rest
    }
  }
}

export function subscribeEvents(fn: Listener): () => void {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}
