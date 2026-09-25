interface StartTerminalHeartbeatParams {
  heartbeat: () => Promise<void>;
  intervalMs?: number;
  /** Called with the current consecutive-failure count after each failure. */
  onFailure?: (consecutiveFailures: number) => void;
  /** Called after a failure streak ends with a successful beat. */
  onRecover?: () => void;
}

const DEFAULT_HEARTBEAT_INTERVAL_MS = 60_000;
// Back off progressively when beats keep failing so an unregistered or
// offline terminal doesn't hammer the API every minute forever.
const MAX_BACKOFF_MULTIPLIER = 8;

export function startTerminalHeartbeat({
  heartbeat,
  intervalMs = DEFAULT_HEARTBEAT_INTERVAL_MS,
  onFailure,
  onRecover,
}: StartTerminalHeartbeatParams): () => void {
  let stopped = false;
  let inFlight = false;
  let consecutiveFailures = 0;
  let currentInterval = intervalMs;
  let intervalId = window.setInterval(tick, currentInterval);

  // Recreate the interval only when the backoff actually changes so the
  // success path keeps its exact cadence.
  const applyBackoff = (): void => {
    const next = Math.min(
      intervalMs * MAX_BACKOFF_MULTIPLIER,
      intervalMs * 2 ** Math.min(consecutiveFailures, 3)
    );
    if (next === currentInterval) return;
    currentInterval = next;
    window.clearInterval(intervalId);
    intervalId = window.setInterval(tick, currentInterval);
  };

  function tick(): void {
    if (stopped || inFlight) return;
    inFlight = true;
    heartbeat()
      .then(() => {
        if (consecutiveFailures > 0) {
          consecutiveFailures = 0;
          onRecover?.();
        }
      })
      .catch(() => {
        if (stopped) return;
        consecutiveFailures += 1;
        onFailure?.(consecutiveFailures);
        applyBackoff();
      })
      .finally(() => {
        inFlight = false;
      });
  }

  tick();

  return () => {
    stopped = true;
    window.clearInterval(intervalId);
  };
}
