import { useEffect, useState } from 'react';

const ACTIVITY = ['mousemove', 'keydown', 'click', 'scroll', 'touchstart'] as const;

/** Returns true after `idleMs` of no user activity. */
export function useIdle(idleMs = 60_000): boolean {
  const [idle, setIdle] = useState(false);

  useEffect(() => {
    let timer: ReturnType<typeof setTimeout> | null = null;
    const reset = (): void => {
      setIdle(false);
      if (timer) clearTimeout(timer);
      timer = setTimeout(() => setIdle(true), idleMs);
    };
    reset();
    for (const ev of ACTIVITY) window.addEventListener(ev, reset, { passive: true });
    return (): void => {
      if (timer) clearTimeout(timer);
      for (const ev of ACTIVITY) window.removeEventListener(ev, reset);
    };
  }, [idleMs]);

  return idle;
}
