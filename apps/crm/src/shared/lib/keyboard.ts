import type { KeyboardEvent } from 'react';

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
