import { useLayoutEffect } from 'react';

/**
 * Make a textarea grow as the user types, capped at `maxRows`. Pass the
 * textarea ref and the current value.
 */
export function useAutosize(
  ref: React.RefObject<HTMLTextAreaElement | null>,
  value: string,
  maxRows = 6
): void {
  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    // Reset to single-row height first so scrollHeight reflects only the
    // current content, then grow up to the cap.
    el.style.height = 'auto';
    const computed = window.getComputedStyle(el);
    const parsedLineHeight = parseFloat(computed.lineHeight);
    const lineHeight =
      Number.isFinite(parsedLineHeight) && parsedLineHeight > 0 ? parsedLineHeight : 20;
    const padTop = parseFloat(computed.paddingTop) || 0;
    const padBottom = parseFloat(computed.paddingBottom) || 0;
    const minHeight = lineHeight + padTop + padBottom;
    const maxHeight = lineHeight * maxRows + padTop + padBottom;
    const next = Math.min(Math.max(el.scrollHeight, minHeight), maxHeight);
    el.style.height = `${next}px`;
  }, [ref, value, maxRows]);
}
