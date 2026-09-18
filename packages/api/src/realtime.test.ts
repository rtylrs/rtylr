import { describe, it, expect, vi, beforeEach } from 'vitest';
import { RealtimeClient } from './realtime';

describe('RealtimeClient', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('subscribes and handles incoming WebSocket messages', () => {
    let mockWsInstance: any = null;
    const capture = (ws: any): void => {
      mockWsInstance = ws;
    };

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
        capture(this);
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
    const capture = (ws: any): void => {
      mockWsInstance = ws;
    };

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
        capture(this);
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

  describe('connection state', () => {
    function stubSocket() {
      let instance: any = null;
      const capture = (ws: any): void => {
        instance = ws;
      };
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
          capture(this);
        }
      }
      vi.stubGlobal('WebSocket', MockWebSocket);
      return () => instance;
    }

    it('reports disconnected before anything has tried to connect', () => {
      stubSocket();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
      });

      expect(client.getStatus()).toBe('disconnected');
      client.close();
    });

    it('hands a new listener the current status immediately, not on next change', () => {
      stubSocket();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
      });

      const seen: string[] = [];
      client.subscribeStatus((s) => seen.push(s));

      expect(seen).toEqual(['disconnected']);
      client.close();
    });

    it('moves to connected when the socket opens and back on close', async () => {
      const socket = stubSocket();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
      });

      const seen: string[] = [];
      client.subscribeStatus((s) => seen.push(s));
      client.subscribe('project:1', vi.fn());

      await vi.waitFor(() => expect(socket()).not.toBeNull());
      expect(seen).toContain('connecting');

      socket().readyState = 1;
      socket().onopen?.();
      expect(client.getStatus()).toBe('connected');
      expect(seen).toContain('connected');

      socket().onclose?.();
      expect(client.getStatus()).toBe('reconnecting');
      expect(seen).toContain('reconnecting');

      client.close();
    });

    it('reports disconnected after an explicit close, never reconnecting', async () => {
      const socket = stubSocket();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
      });
      client.subscribe('project:1', vi.fn());
      await vi.waitFor(() => expect(socket()).not.toBeNull());
      socket().onopen?.();

      client.close();

      expect(client.getStatus()).toBe('disconnected');
    });

    it('stops notifying a listener that unsubscribed', async () => {
      const socket = stubSocket();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
      });

      const seen: string[] = [];
      const unsub = client.subscribeStatus((s) => seen.push(s));
      unsub();
      client.subscribe('project:1', vi.fn());
      await vi.waitFor(() => expect(socket()).not.toBeNull());
      socket().onopen?.();

      expect(seen).toEqual(['disconnected']);
      client.close();
    });

    it('does not let one throwing listener stop the others', () => {
      stubSocket();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
      });

      const second = vi.fn();
      client.subscribeStatus(() => {
        throw new Error('listener blew up');
      });
      client.subscribeStatus(second);
      client.subscribe('project:1', vi.fn());

      expect(second).toHaveBeenCalled();
      client.close();
    });
  });

  describe('onDiagnostic', () => {
    function stubSocket() {
      let instance: any = null;
      const capture = (ws: any): void => {
        instance = ws;
      };
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
          capture(this);
        }
      }
      vi.stubGlobal('WebSocket', MockWebSocket);
      return () => instance;
    }

    it('reports a dropped malformed message without affecting delivery of the next good one', async () => {
      const socket = stubSocket();
      const onDiagnostic = vi.fn();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
        onDiagnostic,
      });
      const handler = vi.fn();
      client.subscribe('project:1', handler);
      await vi.waitFor(() => expect(socket()).not.toBeNull());
      socket().readyState = 1;
      socket().onopen?.();

      socket().onmessage({ data: 'not json{{{' });
      expect(onDiagnostic).toHaveBeenCalledWith({ type: 'malformed_message' });
      expect(handler).not.toHaveBeenCalled();

      socket().onmessage({
        data: JSON.stringify({ op: 'event', event: { topic: 'project:1', type: 'updated' } }),
      });
      expect(handler).toHaveBeenCalledWith({ topic: 'project:1', type: 'updated' });

      client.close();
    });

    it('does not fire for the first two reconnect failures', async () => {
      const socket = stubSocket();
      const onDiagnostic = vi.fn();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
        onDiagnostic,
      });
      client.subscribe('project:1', vi.fn());
      await vi.waitFor(() => expect(socket()).not.toBeNull());
      socket().onopen?.();

      socket().onclose?.();
      expect(onDiagnostic).not.toHaveBeenCalled();

      client.close();
    });

    it('fires once when reconnect failures cross the repeated-failure threshold', async () => {
      // scheduleReconnect no-ops while its own backoff timer is still pending,
      // so getting three real failures counted needs that timer to actually
      // fire (spawning the next socket) between each close - fake time lets
      // that happen without the test taking 30+ real seconds.
      vi.useFakeTimers();
      try {
        const socket = stubSocket();
        const onDiagnostic = vi.fn();
        const client = new RealtimeClient({
          url: 'wss://test.voxire.com/ws',
          getTicket: async () => 'ticket',
          onDiagnostic,
        });
        client.subscribe('project:1', vi.fn());
        await vi.advanceTimersByTimeAsync(0);
        expect(socket()).not.toBeNull();

        socket().onclose?.(); // failure 1/3
        expect(onDiagnostic).not.toHaveBeenCalled();
        await vi.advanceTimersByTimeAsync(30_000); // let backoff fire -> new socket

        socket().onclose?.(); // failure 2/3
        expect(onDiagnostic).not.toHaveBeenCalled();
        await vi.advanceTimersByTimeAsync(30_000);

        socket().onclose?.(); // failure 3/3 - crosses the threshold
        expect(onDiagnostic).toHaveBeenCalledWith({
          type: 'repeated_reconnect_failures',
          attempt: 3,
        });
        expect(onDiagnostic).toHaveBeenCalledTimes(1);

        client.close();
      } finally {
        vi.useRealTimers();
      }
    });

    it('never calls onDiagnostic when the caller does not supply one', async () => {
      const socket = stubSocket();
      const client = new RealtimeClient({
        url: 'wss://test.voxire.com/ws',
        getTicket: async () => 'ticket',
      });
      client.subscribe('project:1', vi.fn());
      await vi.waitFor(() => expect(socket()).not.toBeNull());

      expect(() => {
        socket().onmessage({ data: 'not json' });
        socket().onclose?.();
        socket().onclose?.();
        socket().onclose?.();
      }).not.toThrow();

      client.close();
    });
  });
});
