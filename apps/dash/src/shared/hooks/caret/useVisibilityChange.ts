import { useEffect, useState } from 'react';

/** Tracks the page's document.visibilityState. */
export function useVisibilityChange(): DocumentVisibilityState {
  const [state, setState] = useState<DocumentVisibilityState>(() =>
    typeof document === 'undefined' ? 'visible' : document.visibilityState
  );

  useEffect(() => {
    const onChange = (): void => setState(document.visibilityState);
    document.addEventListener('visibilitychange', onChange);
    return (): void => document.removeEventListener('visibilitychange', onChange);
  }, []);

  return state;
}
