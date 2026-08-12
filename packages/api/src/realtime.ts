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

interface ClientOptions {
  /** Full ws URL, e.g. `wss://api.voxire.com/flow/ws`. */
  url: string;
  /** Function returning the current auth token. Called on every (re)connect. */
  getToken: () => string | null | undefined;
}

interface TopicSub {
  topic: string;
  handlers: Set<EventHandler>;
}

const MIN_BACKOFF_MS = 500;
const MAX_BACKOFF_MS = 30_000;

export class RealtimeClient {
  private url: string;
  private getToken: () => string | null | undefined;
  private ws: WebSocket | null = null;
  private subs = new Map<string, TopicSub>();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private closed = false;
  private pingInterval: ReturnType<typeof setInterval> | null = null;

  constructor(opts: ClientOptions) {
    this.url = opts.url;
    this.getToken = opts.getToken;
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

    const token = this.getToken();
    if (!token) {
      // No auth yet; retry shortly - the user is probably still loading.
      this.scheduleReconnect();
      return;
    }

    const url = `${this.url}?token=${encodeURIComponent(token)}`;
    let ws: WebSocket;
    try {
      ws = new WebSocket(url);
    } catch {
      this.scheduleReconnect();
      return;
    }
    this.ws = ws;

    ws.onopen = (): void => {
      this.reconnectAttempts = 0;
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
        // Malformed payloads: ignore.
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
      if (this.closed) return;
      this.scheduleReconnect();
    };
  }

  private scheduleReconnect(): void {
    if (this.closed || this.reconnectTimer) return;
    const attempt = this.reconnectAttempts++;
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
