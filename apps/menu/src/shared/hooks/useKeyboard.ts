import { useEffect } from 'react';

type KeyHandler = (event: KeyboardEvent) => void;

interface UseKeyboardOptions {
  enabled?: boolean;
  target?: EventTarget | null;
}

/**
 * Attach a keydown listener for one or more key names.
 * The handler fires when the current key matches any entry in the `keys` list.
 * Pass `enabled: false` to temporarily disable without unmounting.
 */
export function useKeyboard(
  keys: string | string[],
  handler: KeyHandler,
  options: UseKeyboardOptions = {}
): void {
  const { enabled = true, target } = options;

  useEffect(() => {
    if (!enabled) return;

    const keyList = Array.isArray(keys) ? keys : [keys];
    const element = target ?? window;

    const onKeyDown = (event: KeyboardEvent) => {
      if (keyList.some((k) => k.toLowerCase() === event.key.toLowerCase())) {
        handler(event);
      }
    };

    element.addEventListener('keydown', onKeyDown as EventListener);
    return () => element.removeEventListener('keydown', onKeyDown as EventListener);
  }, [keys, handler, enabled, target]);
}
