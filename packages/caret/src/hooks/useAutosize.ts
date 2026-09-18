import { useLayoutEffect } from 'react';

/**
 * Make a textarea grow as the user types, capped at `maxRows`. The
 * textarea's `line-height` and vertical padding must match the constants
 * below - we don't read computed styles because flex parents can stretch
 * the element to fill remaining space, making scrollHeight unreliable.
 */
export function useAutosize(
  ref: React.RefObject<HTMLTextAreaElement | null>,
  value: string,
  maxRows = 6
): void {
  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;

    const lineHeight = 20;
    const padY = 8; // 4 top + 4 bottom
    const minHeight = lineHeight + padY;
    const maxHeight = lineHeight * maxRows + padY;

    if (!value) {
      el.style.height = `${minHeight}px`;
      return;
    }

    // Collapse first so scrollHeight measures only the content.
    el.style.height = `${minHeight}px`;
    const grow = Math.max(minHeight, Math.min(el.scrollHeight, maxHeight));
    el.style.height = `${grow}px`;
  }, [ref, value, maxRows]);
}
