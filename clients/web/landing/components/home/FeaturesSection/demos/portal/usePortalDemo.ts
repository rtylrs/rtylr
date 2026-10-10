import { useEffect, useState } from 'react';
import { useReducedMotion } from 'motion/react';

/**
 * Steps through 0..count-1 every `intervalMs` while the demo is on screen,
 * restarting from 0 each time it becomes active. With reduced motion it
 * holds on `restingStep`, a representative frame.
 */
export function useDemoStep(
  isActive: boolean,
  count: number,
  intervalMs: number,
  restingStep = count - 1,
): number {
  const reduceMotion = useReducedMotion();
  const [step, setStep] = useState(0);

  useEffect(() => {
    if (!isActive || reduceMotion) return;
    setStep(0);
    const timer = setInterval(() => setStep((s) => (s + 1) % count), intervalMs);
    return () => clearInterval(timer);
  }, [isActive, reduceMotion, count, intervalMs]);

  return reduceMotion ? restingStep : step;
}

/** Eases a number from 0 to `target` when the demo becomes active. */
export function useCountUp(target: number, isActive: boolean, durationMs = 1100): number {
  const reduceMotion = useReducedMotion();
  const [value, setValue] = useState(target);

  useEffect(() => {
    if (!isActive || reduceMotion) {
      setValue(target);
      return;
    }
    let raf = 0;
    const start = performance.now();
    const tick = (now: number): void => {
      const p = Math.min(1, (now - start) / durationMs);
      setValue(target * (1 - Math.pow(1 - p, 3)));
      if (p < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [target, isActive, reduceMotion, durationMs]);

  return value;
}

export const money = (n: number, digits = 0): string =>
  `$${n.toLocaleString('en-US', { minimumFractionDigits: digits, maximumFractionDigits: digits })}`;
