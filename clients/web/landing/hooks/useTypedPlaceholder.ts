import { useEffect, useState } from 'react';
import { useReducedMotion } from 'motion/react';

const TYPE_MS = 38;
const ERASE_MS = 18;
const HOLD_MS = 1800;

/**
 * Types each example out, holds it, erases it, then moves to the next one.
 * Returns `fallback` (static text) while paused, before hydration, or when the
 * reader prefers reduced motion, so the field is never blank or jumpy.
 */
export function useTypedPlaceholder(
  examples: string[],
  fallback: string,
  paused: boolean,
): string {
  const reduceMotion = useReducedMotion();
  // Server HTML has the static fallback; typing starts only after mount so
  // the first client render matches it.
  const [mounted, setMounted] = useState(false);
  const [index, setIndex] = useState(0);
  const [length, setLength] = useState(0);
  const [erasing, setErasing] = useState(false);

  useEffect(() => setMounted(true), []);

  const active = mounted && !paused && !reduceMotion && examples.length > 0;
  const current = examples[index % Math.max(examples.length, 1)] ?? '';

  useEffect(() => {
    if (!active) return;

    let delay = erasing ? ERASE_MS : TYPE_MS;
    if (!erasing && length === current.length) delay = HOLD_MS;

    const timer = setTimeout(() => {
      if (!erasing && length < current.length) {
        setLength(length + 1);
      } else if (!erasing) {
        setErasing(true);
      } else if (length > 0) {
        setLength(length - 1);
      } else {
        setErasing(false);
        setIndex((i) => (i + 1) % examples.length);
      }
    }, delay);

    return () => clearTimeout(timer);
  }, [active, erasing, length, current, examples.length]);

  if (!active) return fallback;
  return current.slice(0, length);
}
