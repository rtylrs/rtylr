/**
 * Trailing-edge throttle. Returns a function that invokes `fn` at most
 * once per `delay` ms, queuing the latest call's arguments if fired
 * while the cooldown is active.
 */
export function throttle<TArgs extends unknown[]>(
  fn: (...args: TArgs) => void,
  delay = 100
): (...args: TArgs) => void {
  let last = 0;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let pendingArgs: TArgs | null = null;

  const fire = (args: TArgs): void => {
    last = Date.now();
    fn(...args);
    pendingArgs = null;
  };

  return (...args: TArgs): void => {
    const now = Date.now();
    const remaining = delay - (now - last);
    if (remaining <= 0) {
      if (timer) {
        clearTimeout(timer);
        timer = null;
      }
      fire(args);
    } else {
      pendingArgs = args;
      if (!timer) {
        timer = setTimeout(() => {
          timer = null;
          if (pendingArgs) fire(pendingArgs);
        }, remaining);
      }
    }
  };
}
