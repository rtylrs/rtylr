import { useEffect, useRef } from 'react';

/** Returns a ref that always points at the latest value. */
export function useLatest<T>(value: T): React.RefObject<T> {
  const ref = useRef(value);
  useEffect(() => {
    ref.current = value;
  }, [value]);
  return ref;
}
