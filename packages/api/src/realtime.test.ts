import { describe, it, expect, vi, beforeEach } from 'vitest';
import { RealtimeClient } from './realtime';

describe('RealtimeClient', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('subscribes and handles incoming WebSocket messages', () => {
    let mockWsInstance: any = null;

    class MockWebSocket {
      static CONNECTING = 0;
      static OPEN = 1;
      static CLOSING = 2;
      static CLOSED = 3;

      readyState = 0; // CONNECTING initially
      send = vi.fn();
      close = vi.fn();
      onopen: any;
      onmessage: any;
      onerror: any;
      onclose: any;
      constructor() {
        mockWsInstance = this;
      }
    }

    vi.stubGlobal('WebSocket', MockWebSocket);

    const client = new RealtimeClient({
      url: 'wss://test.voxire.com/ws',
      getToken: () => 'valid-token',
    });

    const handler = vi.fn();
    const unsub = client.subscribe('project:123', handler);

    expect(handler).not.toHaveBeenCalled();

    // Trigger onopen
    mockWsInstance.readyState = 1; // OPEN
    mockWsInstance.onopen();
    expect(mockWsInstance.send).toHaveBeenCalledWith(
      JSON.stringify({ op: 'subscribe', topics: ['project:123'] })
    );

    // Trigger non-string message (should be safely ignored)
    mockWsInstance.onmessage({ data: new ArrayBuffer(8) });
    expect(handler).not.toHaveBeenCalled();

    // Trigger valid event message
    mockWsInstance.onmessage({
      data: JSON.stringify({
        op: 'event',
        event: { topic: 'project:123', type: 'updated', entity_id: 1 },
      }),
    });

    expect(handler).toHaveBeenCalledWith({
      topic: 'project:123',
      type: 'updated',
      entity_id: 1,
    });

    unsub();
    client.close();
  });
});
