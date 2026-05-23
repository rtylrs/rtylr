import { useEffect, useState } from 'react';

/**
 * Returns a tick counter that increments every second.
 * Components use this to trigger elapsed-time recalculations without
 * storing per-order timers - a single counter re-renders all tickets.
 */
export function useKdsTick(): number {
  const [tick, setTick] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setTick((n) => n + 1), 1000);
    return () => clearInterval(id);
  }, []);

  return tick;
}

/** Format elapsed seconds as MM:SS, turning amber at 5 min and red at 10 min. */
export function formatElapsed(seconds: number): { label: string; urgency: 'ok' | 'warn' | 'crit' } {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  const label = `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  const urgency = seconds >= 600 ? 'crit' : seconds >= 300 ? 'warn' : 'ok';
  return { label, urgency };
}

export function elapsedSeconds(createdAt: string): number {
  return Math.max(0, Math.floor((Date.now() - new Date(createdAt).getTime()) / 1000));
}
