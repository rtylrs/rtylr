import { useRef } from 'react';

let counter = 0;

/** Stable per-component opaque ID. Mirrors React's useId for older React. */
export function useId(prefix = 'caret'): string {
  const ref = useRef<string | null>(null);
  if (ref.current === null) ref.current = `${prefix}-${++counter}`;
  return ref.current;
}
