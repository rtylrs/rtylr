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
      constructor(public url: string) {
        mockWsInstance = this;
      }
    }

    vi.stubGlobal('WebSocket', MockWebSocket);

    const client = new RealtimeClient({
      url: 'wss://test.voxire.com/ws',
      getTicket: async () => 'one-time-ticket',
    });

    const handler = vi.fn();
    const unsub = client.subscribe('project:123', handler);

    return vi
      .waitFor(() => {
        expect(mockWsInstance.url).toBe('wss://test.voxire.com/ws?ticket=one-time-ticket');
        expect(mockWsInstance.url).not.toContain('token=');
      })
      .then(() => {
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

  it('does not open a socket when the ticket exchange has no authenticated session', async () => {
    const webSocket = vi.fn();
    vi.stubGlobal('WebSocket', webSocket);
    const getTicket = vi.fn().mockResolvedValue(null);
    const client = new RealtimeClient({
      url: 'wss://test.voxire.com/ws',
      getTicket,
    });

    client.subscribe('project:123', vi.fn());
    await vi.waitFor(() => expect(getTicket).toHaveBeenCalled());
    expect(webSocket).not.toHaveBeenCalled();
    client.close();
  });

  it('removes a legacy token from the supplied URL before opening the socket', async () => {
    let mockWsInstance: any = null;

    class MockWebSocket {
      static CONNECTING = 0;
      static OPEN = 1;
      static CLOSING = 2;
      static CLOSED = 3;

      readyState = 0;
      close = vi.fn();
      send = vi.fn();
      onopen: any;
      onmessage: any;
      onerror: any;
      onclose: any;
      constructor(public url: string) {
        mockWsInstance = this;
      }
    }

    vi.stubGlobal('WebSocket', MockWebSocket);
    const client = new RealtimeClient({
      url: 'wss://test.voxire.com/ws?token=reusable-jwt',
      getTicket: async () => 'single-use-ticket',
    });

    client.subscribe('project:123', vi.fn());
    await vi.waitFor(() => expect(mockWsInstance).not.toBeNull());
    expect(mockWsInstance.url).toBe('wss://test.voxire.com/ws?ticket=single-use-ticket');
    expect(mockWsInstance.url).not.toContain('reusable-jwt');
    client.close();
  });
});
