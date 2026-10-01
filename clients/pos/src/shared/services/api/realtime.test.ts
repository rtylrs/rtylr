import { describe, expect, it, vi } from 'vitest';

const realtimeMocks = vi.hoisted(() => {
  type Event = { topic: string; type: string };
  type EventHandler = (event: Event) => void;

  class MockRealtimeClient {
    readonly subscriptions = new Map<string, Set<EventHandler>>();

    constructor(readonly options: { url: string }) {}

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
      this.subscriptions.clear();
    }
  }

  const clients: MockRealtimeClient[] = [];

  return { MockRealtimeClient, clients };
});

vi.mock('@voxire/api', () => ({
  RealtimeClient: class extends realtimeMocks.MockRealtimeClient {
    constructor(options: { url: string }) {
      super(options);
      realtimeMocks.clients.push(this);
    }
  },
  createOneKitFetch: () => vi.fn(),
}));

vi.mock('../server/serverConfig', () => ({
  resolveApiBaseUrl: () => 'https://api.rtylr.example/',
  resolveWsBaseUrl: () => 'wss://api.rtylr.example',
}));

import { posRealtime } from './realtime';

describe('posRealtime', () => {
  it('delivers events and closes a topic after its final listener leaves', () => {
    const handler = vi.fn();
    const unsubscribe = posRealtime.subscribe('owner:1', handler);

    expect(realtimeMocks.clients).toHaveLength(1);
    expect(realtimeMocks.clients[0]?.options.url).toBe(
      'wss://api.rtylr.example/pos/ws'
    );
    expect(realtimeMocks.clients[0]?.subscriptions.get('owner:1')?.size).toBe(1);

    const activeHandler = Array.from(
      realtimeMocks.clients[0]?.subscriptions.get('owner:1') ?? []
    )[0];
    activeHandler?.({ topic: 'owner:1', type: 'pos:order:voided' });
    expect(handler).toHaveBeenCalledWith({
      topic: 'owner:1',
      type: 'pos:order:voided',
    });

    unsubscribe();
    expect(realtimeMocks.clients[0]?.subscriptions.has('owner:1')).toBe(false);
  });
});
