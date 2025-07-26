import { useState, useEffect, useCallback } from 'react';

function getWindowDimensions() {
  if (typeof window === 'undefined') {
    return { width: 0, height: 0 };
  }
  const { innerWidth: width, innerHeight: height } = window;
  return {
    width,
    height,
  };
}

function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): T {
  let inThrottle: boolean;
  return ((...args: any[]) => {
    if (!inThrottle) {
      func.apply(null, args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  }) as T;
}

export function useWindowDimensions() {
  const [windowDimensions, setWindowDimensions] = useState(
    getWindowDimensions()
  );
  const [isMobile, setIsMobile] = useState(windowDimensions.width <= 768);

  const handleResize = useCallback(
    throttle(() => {
      const dimensions = getWindowDimensions();
      setWindowDimensions(dimensions);
      setIsMobile(dimensions.width <= 768);
    }, 100),
    []
  );

  useEffect(() => {
    if (typeof window === 'undefined') return;

    window.addEventListener('resize', handleResize, { passive: true });
    return () => window.removeEventListener('resize', handleResize);
  }, [handleResize]);

  return { ...windowDimensions, isMobile };
}
