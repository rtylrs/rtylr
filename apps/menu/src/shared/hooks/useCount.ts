import { useState } from 'react';

interface UseCountReturn {
  count: number;
  increment: () => void;
  decrement: () => void;
  reset: () => void;
  setCount: (value: number) => void;
}

/**
 * A simple counter hook with increment, decrement, and reset.
 * An optional min and max clamp the count to a bounded range.
 */
export function useCount(
  initial = 0,
  options: { min?: number; max?: number } = {}
): UseCountReturn {
  const { min, max } = options;
  const [count, setRaw] = useState(initial);

  function clamp(value: number): number {
    let result = value;
    if (min !== undefined) result = Math.max(result, min);
    if (max !== undefined) result = Math.min(result, max);
    return result;
  }

  return {
    count,
    increment: () => setRaw((c) => clamp(c + 1)),
    decrement: () => setRaw((c) => clamp(c - 1)),
    reset: () => setRaw(clamp(initial)),
    setCount: (value: number) => setRaw(clamp(value)),
  };
}
