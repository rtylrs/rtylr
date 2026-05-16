import { useEffect, useState } from 'react';

export interface WindowSize {
  width: number;
  height: number;
}

/** Reactive window size, debounced via rAF. */
export function useWindowSize(): WindowSize {
  const [size, setSize] = useState<WindowSize>(() => ({
    width: typeof window !== 'undefined' ? window.innerWidth : 0,
    height: typeof window !== 'undefined' ? window.innerHeight : 0,
  }));

  useEffect(() => {
    let raf: number | null = null;
    const onResize = (): void => {
      if (raf != null) return;
      raf = requestAnimationFrame(() => {
        raf = null;
        setSize({ width: window.innerWidth, height: window.innerHeight });
      });
    };
    window.addEventListener('resize', onResize);
    return (): void => {
      if (raf != null) cancelAnimationFrame(raf);
      window.removeEventListener('resize', onResize);
    };
  }, []);

  return size;
}
