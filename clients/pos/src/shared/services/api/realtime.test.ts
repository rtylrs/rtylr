import { describe, expect, it, vi } from "vitest";

const realtimeMocks = vi.hoisted(() => {
  type Event = { topic: string; type: string };
  type EventHandler = (event: Event) => void;
  type ServerState = {
    mode: "hosted" | "self_host";
    customApiUrl: string | null;
  };
  type ServerListener = (
    state: ServerState,
    previousState: ServerState,
  ) => void;

  class MockRealtimeClient {
    readonly subscriptions = new Map<string, Set<EventHandler>>();
    closed = false;

    constructor(readonly options: { url: string }) {
      clients.push(this);
    }

    subscribe(topic: string, handler: EventHandler): () => void {
      const handlers = this.subscriptions.get(topic) ?? new Set<EventHandler>();
      handlers.add(handler);
      this.subscriptions.set(topic, handlers);
      return (): void => {
        handlers.delete(handler);
        if (handlers.size === 0) this.subscriptions.delete(topic);
      };
    }

    close(): void {
      this.closed = true;
      this.subscriptions.clear();
    }
  }

  const clients: MockRealtimeClient[] = [];
  const listeners = new Set<ServerListener>();
  let serverState: ServerState = { mode: "hosted", customApiUrl: null };
  let wsBaseUrl = "wss://server-a.example";

  return {
    MockRealtimeClient,
    clients,
    getWsBaseUrl: (): string => wsBaseUrl,
    serverStore: {
      subscribe: (listener: ServerListener): (() => void) => {
        listeners.add(listener);
        return (): void => {
          listeners.delete(listener);
        };
      },
    },
    setServer: (next: ServerState): void => {
      const previous = serverState;
      serverState = next;
      for (const listener of listeners) listener(serverState, previous);
      if (next.mode === "self_host" && next.customApiUrl) {
        wsBaseUrl = next.customApiUrl.replace(/^http/, "ws").replace(/\/$/, "");
      }
    },
  };
});

vi.mock("@voxire/api", () => ({
  RealtimeClient: realtimeMocks.MockRealtimeClient,
  createOneKitFetch: () => vi.fn(),
}));

vi.mock("../../stores/serverStore", () => ({
  useServerStore: realtimeMocks.serverStore,
}));

vi.mock("../server/serverConfig", () => ({
  resolveApiBaseUrl: () => "https://server-a.example/",
  resolveWsBaseUrl: realtimeMocks.getWsBaseUrl,
}));

import { posRealtime } from "./realtime";

describe("posRealtime", () => {
  it("reconnects active subscriptions to the newly selected server", async () => {
    const handler = vi.fn();
    posRealtime.subscribe("owner:1", handler);

    expect(realtimeMocks.clients).toHaveLength(1);
    expect(realtimeMocks.clients[0]?.options.url).toBe(
      "wss://server-a.example/pos/ws",
    );

    realtimeMocks.setServer({
      mode: "self_host",
      customApiUrl: "https://server-b.example/",
    });
    await Promise.resolve();

    expect(realtimeMocks.clients).toHaveLength(2);
    expect(realtimeMocks.clients[0]?.closed).toBe(true);
    expect(realtimeMocks.clients[1]?.options.url).toBe(
      "wss://server-b.example/pos/ws",
    );
    expect(realtimeMocks.clients[1]?.subscriptions.has("owner:1")).toBe(true);

    const activeHandler = Array.from(
      realtimeMocks.clients[1]?.subscriptions.get("owner:1") ?? [],
    )[0];
    activeHandler?.({ topic: "owner:1", type: "pos:order:voided" });
    expect(handler).toHaveBeenCalledWith({
      topic: "owner:1",
      type: "pos:order:voided",
    });
  });
});
