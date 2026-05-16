import { useEffect, useRef } from 'react';

/** Returns a ref that always tracks the latest value. */
export function useRefValue<T>(value: T): React.RefObject<T> {
  const ref = useRef(value);
  useEffect(() => {
    ref.current = value;
  }, [value]);
  return ref;
}
