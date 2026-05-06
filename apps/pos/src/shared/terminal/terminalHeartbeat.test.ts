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
});
