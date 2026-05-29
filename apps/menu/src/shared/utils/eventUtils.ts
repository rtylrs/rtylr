/**
 * DOM and custom event helpers for the Menu app.
 */

/** Dispatch a custom event on the document. */
export function dispatchCustomEvent<T = unknown>(name: string, detail?: T): void {
  document.dispatchEvent(new CustomEvent(name, { detail, bubbles: true }));
}

/** Return true if the keyboard event matches the given key (case-insensitive). */
export function isKey(event: KeyboardEvent, key: string): boolean {
  return event.key.toLowerCase() === key.toLowerCase();
}

/** Return true if the event target is an interactive element (input, button, etc). */
export function isInteractiveTarget(event: Event): boolean {
  const tag = (event.target as HTMLElement)?.tagName?.toLowerCase();
  return ['input', 'textarea', 'select', 'button', 'a'].includes(tag);
}

/** Stop event propagation and prevent default. */
export function stopEvent(event: Event): void {
  event.stopPropagation();
  event.preventDefault();
}

/** Return true if the mouse event is a primary (left) button click. */
export function isPrimaryClick(event: MouseEvent): boolean {
  return event.button === 0;
}

/** Return the scroll position of the window as {x, y}. */
export function getScrollPosition(): { x: number; y: number } {
  return { x: window.scrollX, y: window.scrollY };
}

/** Scroll the window to the top smoothly. */
export function scrollToTop(): void {
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

/** Return the key from a keyboard event, normalising Enter and Space. */
export function normaliseKey(event: KeyboardEvent): string {
  return event.key;
}
