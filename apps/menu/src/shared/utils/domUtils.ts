/**
 * DOM utility helpers for the menu app.
 * These functions are designed to be browser-safe and unit-testable via mocks.
 */

/**
 * Scrolls the element with the given id into view.
 * Does nothing when no matching element is found.
 */
export function scrollToId(id: string, behavior: ScrollBehavior = 'smooth'): void {
  const el = document.getElementById(id);
  if (el) {
    el.scrollIntoView({ behavior, block: 'start' });
  }
}

/**
 * Returns the data-* attribute value from a click target element.
 * Walks up the DOM tree until the attribute is found or the root is reached.
 */
export function dataAttrFromClick(event: MouseEvent, attr: string): string | null {
  let target = event.target as HTMLElement | null;
  while (target) {
    const value = target.dataset[attr];
    if (value !== undefined) return value;
    target = target.parentElement;
  }
  return null;
}

/**
 * Returns true when the element is scrolled to the bottom within the given
 * threshold (in pixels).
 */
export function isScrolledToBottom(el: HTMLElement, threshold = 10): boolean {
  return el.scrollHeight - el.scrollTop - el.clientHeight <= threshold;
}

/**
 * Copies the given text to the clipboard.
 * Returns a Promise that resolves to true on success or false on failure.
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}

/**
 * Traps focus within the given element by preventing Tab from escaping.
 * Returns a cleanup function that removes the listener.
 */
export function trapFocus(container: HTMLElement): () => void {
  const focusable = container.querySelectorAll<HTMLElement>(
    'a, button, input, textarea, select, [tabindex]:not([tabindex="-1"])'
  );

  const first = focusable[0];
  const last = focusable[focusable.length - 1];

  function onKeyDown(e: KeyboardEvent) {
    if (e.key !== 'Tab') return;
    if (e.shiftKey) {
      if (document.activeElement === first) {
        e.preventDefault();
        last?.focus();
      }
    } else {
      if (document.activeElement === last) {
        e.preventDefault();
        first?.focus();
      }
    }
  }

  container.addEventListener('keydown', onKeyDown);
  return () => container.removeEventListener('keydown', onKeyDown);
}
