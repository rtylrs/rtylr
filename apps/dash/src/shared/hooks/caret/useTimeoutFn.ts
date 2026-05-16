import { useCallback, useEffect, useRef } from 'react';

/** Schedule `fn` to run once after `ms`. Returns `cancel`. */
export function useTimeoutFn(
  fn: () => void,
  ms: number
): { cancel: () => void; restart: () => void } {
  const ref = useRef(fn);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    ref.current = fn;
  }, [fn]);

  const cancel = useCallback(() => {
    if (timer.current) {
      clearTimeout(timer.current);
      timer.current = null;
    }
  }, []);

  const restart = useCallback(() => {
    cancel();
    timer.current = setTimeout(() => ref.current(), ms);
  }, [cancel, ms]);

  useEffect(() => {
    restart();
    return cancel;
  }, [restart, cancel]);

  return { cancel, restart };
}
