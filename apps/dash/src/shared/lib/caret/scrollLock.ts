/**
 * Temporarily disable body scroll. Used by the approval dialog so the
 * background doesn't drift when keyboard focus is trapped.
 */
let count = 0;
let prev = '';

export function lockScroll(): void {
  if (typeof document === 'undefined') return;
  if (count === 0) {
    prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
  }
  count++;
}

export function unlockScroll(): void {
  if (typeof document === 'undefined') return;
  count = Math.max(0, count - 1);
  if (count === 0) {
    document.body.style.overflow = prev;
  }
}
