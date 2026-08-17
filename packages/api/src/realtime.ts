/**
 * Realtime client for the voxire websocket bus.
 *
 * The server at `/flow/ws` accepts a single connection per tab and
 * multiplexes events across topics (`user:<id>`, `project:<id>`, etc.).
 * Events are lightweight invalidation hints: `{ topic, type, entity_id,
 * actor_id }`. The canonical entity data is still fetched via REST - the
 * websocket only tells the client *that* something changed so it knows
 * when to refetch.
 *
 * This module exposes a singleton `realtime` client used by a React hook
 * (`useRealtime`, defined in the consuming app). A single socket is
 * shared by every subscriber so opening three projects concurrently
 * doesn't open three sockets.
 */
export interface RealtimeEvent {
  topic: string;
  type: string;
  /** Affected entity id, 0 if the event is collection-level. */
  entity_id?: number;
  /** The user who caused this event. Consumers skip self-triggered refetches. */
  actor_id?: number;
}

type EventHandler = (e: RealtimeEvent) => void;

/**
 * Non-fatal signal worth a breadcrumb before something user-visible breaks.
 * Both cases are already handled automatically by the client - a malformed
 * frame is dropped, a failing reconnect keeps retrying with backoff - so this
 * is observability only. Emitted once per `repeated_reconnect_failures`
 * streak (not on every retry) to avoid flooding whatever the caller wires it
 * to for as long as the outage lasts.
 */
export type RealtimeDiagnosticEvent =
  | { type: 'malformed_message' }
  | { type: 'repeated_reconnect_failures'; attempt: number };

interface ClientOptions {
  /** Full ws URL, e.g. `wss://api.voxire.com/flow/ws`. */
  url: string;
  /**
   * Exchanges the authenticated REST session for a short-lived, one-time
   * WebSocket ticket. Called before every (re)connect so a failed handshake
   * never reuses a ticket that may already have been consumed.
   */
  getTicket: () => Promise<string | null | undefined>;
  /**
   * Optional hook for reporting non-fatal conditions (a dropped malformed
   * message, a reconnect that keeps failing) to whatever observability the
   * caller wires up. Nothing today calls this if omitted - the client's own
   * recovery behavior is unaffected either way.
   */
  onDiagnostic?: (event: RealtimeDiagnosticEvent) => void;
}

interface TopicSub {
  topic: string;
  handlers: Set<EventHandler>;
}

const MIN_BACKOFF_MS = 500;
const MAX_BACKOFF_MS = 30_000;
/** Consecutive reconnect failures before `repeated_reconnect_failures` fires. */
const REPEATED_RECONNECT_THRESHOLD = 3;

/**
 * Connection state, for surfaces that must tell the user their view may be
 * stale. `reconnecting` is distinct from `disconnected`: the former is retrying
 * with backoff, the latter followed an explicit close and will never recover on
 * its own.
 */
export type RealtimeStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

type StatusHandler = (status: RealtimeStatus) => void;

export class RealtimeClient {
  private url: string;
  private getTicket: () => Promise<string | null | undefined>;
  private onDiagnostic?: (event: RealtimeDiagnosticEvent) => void;
  private ws: WebSocket | null = null;
  private subs = new Map<string, TopicSub>();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private closed = false;
  private pingInterval: ReturnType<typeof setInterval> | null = null;
  private connectionPromise: Promise<void> | null = null;
  private status: RealtimeStatus = 'disconnected';
  private statusHandlers = new Set<StatusHandler>();

  constructor(opts: ClientOptions) {
    this.url = opts.url;
    this.getTicket = opts.getTicket;
    this.onDiagnostic = opts.onDiagnostic;
  }

  /** Current connection state. */
  getStatus(): RealtimeStatus {
    return this.status;
  }

  /**
   * Observe the connection state. The handler fires immediately with the current
   * status, so a component that mounts mid-connection renders the truth rather
   * than waiting for the next transition. Returns an unsubscribe function.
   */
  subscribeStatus(handler: StatusHandler): () => void {
    this.statusHandlers.add(handler);
    // Guarded like every other notification: the immediate call is the one most
    // likely to throw, because it runs during the caller's own render.
    try {
      handler(this.status);
    } catch {
      // Registration still succeeded; a broken handler is the handler's problem.
    }
    return () => {
      this.statusHandlers.delete(handler);
    };
  }

  private setStatus(next: RealtimeStatus): void {
    if (this.status === next) return;
    this.status = next;
    for (const handler of this.statusHandlers) {
      try {
        handler(next);
      } catch {
        // A listener that throws must not stop the others, or one broken
        // indicator silences every other consumer of this state.
      }
    }
  }

  private notifyDiagnostic(event: RealtimeDiagnosticEvent): void {
    try {
      this.onDiagnostic?.(event);
    } catch {
      // A broken diagnostic hook must not affect message handling or
      // reconnect scheduling - it is observability, not control flow.
    }
  }

  /** Subscribe a handler to `topic`. Returns an unsubscribe function. */
  subscribe(topic: string, handler: EventHandler): () => void {
    let entry = this.subs.get(topic);
    if (!entry) {
      entry = { topic, handlers: new Set() };
      this.subs.set(topic, entry);
      // First subscriber for this topic: tell the server if we're
      // already connected; otherwise the connect path will re-subscribe
      // everything on (re)open.
      this.send({ op: 'subscribe', topics: [topic] });
    }
    entry.handlers.add(handler);

    // Ensure the socket is coming up.
    this.ensureConnected();

    return () => this.unsubscribe(topic, handler);
  }

  /** Gracefully close. Handlers stop receiving events immediately. */
  close(): void {
    this.closed = true;
    this.subs.clear();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
    this.ws?.close();
    this.ws = null;
    // Set directly rather than waiting for onclose: a socket that never opened
    // has no onclose to fire, and the caller has already been told the session
    // is over.
    this.setStatus('disconnected');
  }

  // ─── private ──────────────────────────────────────────────────────────

  private unsubscribe(topic: string, handler: EventHandler): void {
    const entry = this.subs.get(topic);
    if (!entry) return;
    entry.handlers.delete(handler);
    if (entry.handlers.size === 0) {
      this.subs.delete(topic);
      this.send({ op: 'unsubscribe', topics: [topic] });
    }
  }

  private ensureConnected(): void {
    if (this.closed) return;
    if (
      this.ws &&
      (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }

    if (this.connectionPromise) return;

    this.connectionPromise = this.connect().finally(() => {
      this.connectionPromise = null;
    });
  }

  private async connect(): Promise<void> {
    if (!this.closed) {
      this.setStatus(this.reconnectAttempts > 0 ? 'reconnecting' : 'connecting');
    }
    let ticket: string | null | undefined;
    try {
      ticket = await this.getTicket();
    } catch {
      ticket = null;
    }

    if (this.closed) return;
    if (!ticket) {
      // No authenticated session yet, or the ticket exchange failed.
      this.scheduleReconnect();
      return;
    }

    let url: URL;
    try {
      url = new URL(this.url);
      // Never carry forward the legacy reusable-JWT query parameter, even if
      // a caller supplies a URL that still contains one.
      url.searchParams.delete('token');
      url.searchParams.set('ticket', ticket);
    } catch {
      this.scheduleReconnect();
      return;
    }

    let ws: WebSocket;
    try {
      ws = new WebSocket(url.toString());
    } catch {
      this.scheduleReconnect();
      return;
    }
    this.ws = ws;

    ws.onopen = (): void => {
      this.reconnectAttempts = 0;
      this.setStatus('connected');
      // Re-subscribe to everything that was active before the connect.
      const topics = Array.from(this.subs.keys());
      if (topics.length > 0) {
        this.send({ op: 'subscribe', topics });
      }
      // Keep intermediaries from killing idle connections.
      this.pingInterval = setInterval(() => {
        this.send({ op: 'ping' });
      }, 25_000);
    };

    ws.onmessage = (evt: MessageEvent): void => {
      try {
        if (typeof evt.data !== 'string') return;
        const msg = JSON.parse(evt.data);
        if (msg?.op === 'event' && msg.event) {
          const e = msg.event as RealtimeEvent;
          const entry = this.subs.get(e.topic);
          if (entry) {
            for (const h of entry.handlers) h(e);
          }
        }
      } catch {
        // Malformed payloads: dropped, but worth a breadcrumb for whoever's
        // watching before it becomes a user-visible symptom.
        this.notifyDiagnostic({ type: 'malformed_message' });
      }
    };

    ws.onerror = (): void => {
      // onerror fires before onclose; actual reconnect is driven by onclose.
    };

    ws.onclose = (): void => {
      if (this.pingInterval) {
        clearInterval(this.pingInterval);
        this.pingInterval = null;
      }
      this.ws = null;
      if (this.closed) {
        this.setStatus('disconnected');
        return;
      }
      // Retrying with backoff, which is not the same as being disconnected: the
      // caller can keep buffering rather than treating the session as over.
      this.setStatus('reconnecting');
      this.scheduleReconnect();
    };
  }

  private scheduleReconnect(): void {
    if (this.closed || this.reconnectTimer) return;
    const attempt = this.reconnectAttempts++;
    // Fire once on crossing the threshold, not on every retry after - the
    // outage can last minutes and a breadcrumb per backoff cycle would just
    // be noise by the time anyone reads it.
    if (attempt + 1 === REPEATED_RECONNECT_THRESHOLD) {
      this.notifyDiagnostic({ type: 'repeated_reconnect_failures', attempt: attempt + 1 });
    }
    const backoff = Math.min(MAX_BACKOFF_MS, MIN_BACKOFF_MS * Math.pow(2, attempt));
    // Full jitter - helps when many clients bounce at once.
    const delay = Math.floor(Math.random() * backoff);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.ensureConnected();
    }, delay);
  }

  private send(msg: object): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    try {
      this.ws.send(JSON.stringify(msg));
    } catch {
      // Socket died between the readyState check and send; reconnect
      // logic will pick it up.
    }
  }
}

let singleton: RealtimeClient | null = null;

/** Lazy-singleton. First caller wins on opts; later callers get the same instance. */
export function getRealtimeClient(opts: ClientOptions): RealtimeClient {
  if (!singleton) {
    singleton = new RealtimeClient(opts);
  }
  return singleton;
}
