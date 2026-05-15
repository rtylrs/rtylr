import { useLayoutEffect, useRef } from 'react';

/**
 * Make a textarea grow as the user types, capped at `maxRows`. Pass the
 * textarea ref and the current value.
 */
export function useAutosize(
  ref: React.RefObject<HTMLTextAreaElement>,
  value: string,
  maxRows = 8
): void {
  const baseHeightRef = useRef<number | null>(null);

  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (baseHeightRef.current == null) {
      // Measure one-row height once.
      const computed = window.getComputedStyle(el);
      const lineHeight = parseFloat(computed.lineHeight || '20');
      baseHeightRef.current = lineHeight;
    }
    el.style.height = 'auto';
    const maxPx = (baseHeightRef.current ?? 20) * maxRows + 16;
    el.style.height = `${Math.min(el.scrollHeight, maxPx)}px`;
  }, [ref, value, maxRows]);
}
