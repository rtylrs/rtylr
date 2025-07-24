import { useState, useEffect, useCallback } from 'react';

function getWindowDimensions() {
  const { innerWidth: width, innerHeight: height } = window;
  return {
    width,
    height,
  };
}

// Throttle function to limit how often the resize handler runs
function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): T {
  let inThrottle: boolean;
  return ((...args: any[]) => {
    if (!inThrottle) {
      func.apply(this as any, args);
      inThrottle = true;
      setTimeout(() => (inThrottle = false), limit);
    }
  }) as T;
}

export default function useWindowDimensions() {
  const [windowDimensions, setWindowDimensions] = useState(
    getWindowDimensions()
  );
  const [isMobile, setIsMobile] = useState(windowDimensions.width <= 768);

  const handleResize = useCallback(
    throttle(() => {
      const dimensions = getWindowDimensions();
      setWindowDimensions(dimensions);
      setIsMobile(dimensions.width <= 768);
    }, 100), // Throttle to 100ms
    []
  );

  useEffect(() => {
    window.addEventListener('resize', handleResize, { passive: true });
    return () => window.removeEventListener('resize', handleResize);
  }, [handleResize]);

  return { ...windowDimensions, isMobile };
}
