import { describe, expect, it, vi } from 'vitest';
import { startTerminalHeartbeat } from './terminalHeartbeat';

describe('startTerminalHeartbeat', () => {
  it('sends a heartbeat immediately and on each interval', async () => {
    vi.useFakeTimers();
    const heartbeat = vi.fn(async () => undefined);

    const stop = startTerminalHeartbeat({ heartbeat, intervalMs: 1_000 });
    await Promise.resolve();

    expect(heartbeat).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(2_000);

    expect(heartbeat).toHaveBeenCalledTimes(3);
    stop();
    vi.useRealTimers();
  });

  it('does not start another heartbeat while one is still in flight', async () => {
    vi.useFakeTimers();
    let finishHeartbeat: (() => void) | undefined;
    const heartbeat = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          finishHeartbeat = resolve;
        })
    );

    const stop = startTerminalHeartbeat({ heartbeat, intervalMs: 1_000 });
    await Promise.resolve();
    await vi.advanceTimersByTimeAsync(3_000);

    expect(heartbeat).toHaveBeenCalledTimes(1);

    finishHeartbeat?.();
    await Promise.resolve();
    await vi.advanceTimersByTimeAsync(1_000);

    expect(heartbeat).toHaveBeenCalledTimes(2);
    stop();
    vi.useRealTimers();
  });

  it('backs off progressively and reports consecutive failures', async () => {
    vi.useFakeTimers();
    const heartbeat = vi.fn(async () => {
      throw new Error('down');
    });
    const failures: number[] = [];

    const stop = startTerminalHeartbeat({
      heartbeat,
      intervalMs: 1_000,
      onFailure: (n) => failures.push(n),
    });
    await Promise.resolve(); // flush initial rejection

    // Backoff doubles per failure: next beat lands at ~2s, not 1s.
    await vi.advanceTimersByTimeAsync(999);
    expect(heartbeat).toHaveBeenCalledTimes(1);
    expect(failures).toEqual([1]);

    await vi.advanceTimersByTimeAsync(1_001);
    await Promise.resolve();
    expect(heartbeat).toHaveBeenCalledTimes(2);
    expect(failures).toEqual([1, 2]);

    stop();
    vi.useRealTimers();
  });

  it('reports recovery when a beat succeeds after failures', async () => {
    vi.useFakeTimers();
    let failing = true;
    const heartbeat = vi.fn(async () => {
      if (failing) throw new Error('down');
    });
    const onRecover = vi.fn();

    const stop = startTerminalHeartbeat({ heartbeat, intervalMs: 1_000, onRecover });
    await Promise.resolve();

    await vi.advanceTimersByTimeAsync(2_000);
    await Promise.resolve();
    expect(onRecover).not.toHaveBeenCalled();

    failing = false;
    await vi.advanceTimersByTimeAsync(4_000);
    await Promise.resolve();
    expect(onRecover).toHaveBeenCalledOnce();

    stop();
    vi.useRealTimers();
  });
});
