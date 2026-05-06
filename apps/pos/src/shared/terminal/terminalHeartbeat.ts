interface StartTerminalHeartbeatParams {
  heartbeat: () => Promise<void>;
  intervalMs?: number;
}

const DEFAULT_HEARTBEAT_INTERVAL_MS = 60_000;

export function startTerminalHeartbeat({
  heartbeat,
  intervalMs = DEFAULT_HEARTBEAT_INTERVAL_MS,
}: StartTerminalHeartbeatParams): () => void {
  let stopped = false;
  let inFlight = false;

  const tick = (): void => {
    if (stopped || inFlight) return;
    inFlight = true;
    heartbeat()
      .catch(() => {})
      .finally(() => {
        inFlight = false;
      });
  };

  tick();
  const intervalId = window.setInterval(tick, intervalMs);

  return () => {
    stopped = true;
    window.clearInterval(intervalId);
  };
}
