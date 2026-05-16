import { useEffect, useRef } from 'react';

/** Returns a ref that becomes true after mount and false on unmount. */
export function useMounted(): React.RefObject<boolean> {
  const ref = useRef(true);
  useEffect(
    () => () => {
      ref.current = false;
    },
    []
  );
  return ref;
}
