import { useLayoutEffect } from "react";

function positivePixel(value: string, fallback: number): number {
  const parsed = Number.parseFloat(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

/**
 * Make a textarea grow as the user types, capped at `maxRows`.
 * The line box and vertical padding are read from the element so callers can
 * change its size without leaving a stale height behind.
 */
export function useAutosize(
  ref: React.RefObject<HTMLTextAreaElement | null>,
  value: string,
  maxRows = 6,
): void {
  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;

    const computed = window.getComputedStyle(el);
    const lineHeight = positivePixel(computed.lineHeight, 20);
    const padY =
      positivePixel(computed.paddingTop, 0) +
      positivePixel(computed.paddingBottom, 0);
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
  });
}
