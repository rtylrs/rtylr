import { useEffect, useRef } from 'react';

/** Set document.title for the lifetime of the component. Restores on unmount. */
export function useDocumentTitle(title: string): void {
  const previous = useRef<string>('');
  useEffect(() => {
    if (typeof document === 'undefined') return;
    previous.current = document.title;
    document.title = title;
    return (): void => {
      document.title = previous.current;
    };
  }, [title]);
}
