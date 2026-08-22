import { useEffect, type KeyboardEvent } from 'react';

/** Runs a pointer-style action when its focusable owner receives Enter or Space. */
export function activateOnKeyboard(
  event: KeyboardEvent<HTMLElement>,
  onActivate: () => void
): void {
  if (event.target !== event.currentTarget) return;
  if (event.key !== 'Enter' && event.key !== ' ') return;

  event.preventDefault();
  onActivate();
}

/** Closes an open overlay when the user presses Escape. */
export function useEscapeKey(enabled: boolean, onEscape: () => void): void {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: globalThis.KeyboardEvent): void => {
      if (event.key !== 'Escape') return;
      event.preventDefault();
      onEscape();
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [enabled, onEscape]);
}
