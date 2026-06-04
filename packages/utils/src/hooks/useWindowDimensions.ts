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

function throttle<T extends (...args: any[]) => any>(func: T, limit: number): T {
  let inThrottle: boolean;
  return ((...args: any[]) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  }) as T;
}

export function useWindowDimensions() {
  const [windowDimensions, setWindowDimensions] = useState(() => {
    // During SSR, return default values
    if (typeof window === 'undefined') {
      return { width: 0, height: 0 };
    }
    return getWindowDimensions();
  });

  const [isMobile, setIsMobile] = useState(() => {
    // During SSR, return false as default
    if (typeof window === 'undefined') {
      return false;
    }
    const dimensions = getWindowDimensions();
    return dimensions.width <= 768;
  });

  const handleResize = useCallback(() => {
    if (typeof window === 'undefined') return;
    const dimensions = getWindowDimensions();
    setWindowDimensions(dimensions);
    setIsMobile(dimensions.width <= 768);
  }, []);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    // Set initial dimensions after mount to ensure we have the correct values
    const dimensions = getWindowDimensions();
    setWindowDimensions(dimensions);
    setIsMobile(dimensions.width <= 768);

    const throttledResize = throttle(handleResize, 100);
    window.addEventListener('resize', throttledResize, { passive: true });
    return () => window.removeEventListener('resize', throttledResize);
  }, [handleResize]);

  return { ...windowDimensions, isMobile };
}
