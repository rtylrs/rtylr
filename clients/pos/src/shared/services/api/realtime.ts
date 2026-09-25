import {
  RealtimeClient,
  createOneKitFetch,
  type RealtimeEvent,
} from "@voxire/api";
import { useServerStore } from "../../stores/serverStore";
import { resolveApiBaseUrl, resolveWsBaseUrl } from "../server/serverConfig";

let oneKitFetch: ReturnType<typeof createOneKitFetch> | null = null;

export async function getRealtimeTicket(): Promise<string | null> {
  const baseUrl = resolveApiBaseUrl().replace(/\/$/, "");
  try {
    if (!oneKitFetch) {
      oneKitFetch = createOneKitFetch("pos");
    }
    const response = await oneKitFetch(`${baseUrl}/pos/ws-ticket`, {
      method: "POST",
      headers: { Accept: "application/json" },
    });
    if (!response.ok) return null;

    const body: unknown = await response.json();
    if (!body || typeof body !== "object" || !("ticket" in body)) return null;
    const ticket = (body as { ticket?: unknown }).ticket;
    return typeof ticket === "string" && ticket.length > 0 ? ticket : null;
  } catch {
    return null;
  }
}

function resolveWsUrl(): string {
  return `${resolveWsBaseUrl()}/pos/ws`;
}

type RealtimeEventHandler = (event: RealtimeEvent) => void;

interface ActiveTopic {
  handlers: Set<RealtimeEventHandler>;
  unsubscribeClient?: () => void;
}

const activeTopics = new Map<string, ActiveTopic>();
let realtimeClient: RealtimeClient | null = null;
let realtimeUrl: string | null = null;

function attachTopic(topic: string, active: ActiveTopic): void {
  if (active.unsubscribeClient) return;
  const client = ensureRealtimeClient();
  active.unsubscribeClient = client.subscribe(topic, (event) => {
    for (const handler of Array.from(active.handlers)) handler(event);
  });
}

function ensureRealtimeClient(): RealtimeClient {
  const nextUrl = resolveWsUrl();
  if (realtimeClient && realtimeUrl === nextUrl) return realtimeClient;

  realtimeClient?.close();
  realtimeClient = new RealtimeClient({
    url: nextUrl,
    getTicket: getRealtimeTicket,
  });
  realtimeUrl = nextUrl;

  for (const active of activeTopics.values())
    active.unsubscribeClient = undefined;
  for (const [topic, active] of activeTopics) {
    if (active.handlers.size > 0) attachTopic(topic, active);
  }
  return realtimeClient;
}

useServerStore.subscribe((state, previousState) => {
  if (
    !realtimeClient ||
    (state.mode === previousState.mode &&
      state.customApiUrl === previousState.customApiUrl)
  ) {
    return;
  }
  // The persist middleware notifies subscribers before writing localStorage.
  queueMicrotask(() => {
    if (realtimeClient) ensureRealtimeClient();
  });
});

export const posRealtime = {
  subscribe(topic: string, handler: RealtimeEventHandler): () => void {
    let active = activeTopics.get(topic);
    if (!active) {
      active = { handlers: new Set() };
      activeTopics.set(topic, active);
    }
    active.handlers.add(handler);
    attachTopic(topic, active);

    let subscribed = true;
    return (): void => {
      if (!subscribed) return;
      subscribed = false;
      const current = activeTopics.get(topic);
      if (!current) return;
      current.handlers.delete(handler);
      if (current.handlers.size === 0) {
        current.unsubscribeClient?.();
        activeTopics.delete(topic);
      }
    };
  },
};
