import { useState, useEffect, useCallback } from "react";

function getWindowDimensions() {
  if (typeof window === "undefined") {
    return { width: 0, height: 0 };
  }
  const { innerWidth: width, innerHeight: height } = window;
  return {
    width,
    height,
  };
}

type Throttled<T extends (...args: never[]) => unknown> = T & {
  cancel: () => void;
};

function throttle<T extends (...args: never[]) => unknown>(
  func: T,
  limit: number,
): Throttled<T> {
  let inThrottle = false;
  let pendingArgs: Parameters<T> | undefined;
  let timer: ReturnType<typeof setTimeout> | undefined;

  const throttled = ((...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args);
      inThrottle = true;
      timer = setTimeout(() => {
        inThrottle = false;
        if (pendingArgs) {
          const nextArgs = pendingArgs;
          pendingArgs = undefined;
          throttled(...nextArgs);
        }
      }, limit);
      return;
    }

    pendingArgs = args;
  }) as Throttled<T>;

  throttled.cancel = () => {
    if (timer !== undefined) clearTimeout(timer);
    inThrottle = false;
    pendingArgs = undefined;
  };

  return throttled;
}

export function useWindowDimensions() {
  const [windowDimensions, setWindowDimensions] = useState(() => {
    // During SSR, return default values
    if (typeof window === "undefined") {
      return { width: 0, height: 0 };
    }
    return getWindowDimensions();
  });

  const [isMobile, setIsMobile] = useState(() => {
    // During SSR, return false as default
    if (typeof window === "undefined") {
      return false;
    }
    const dimensions = getWindowDimensions();
    return dimensions.width <= 768;
  });

  const handleResize = useCallback(() => {
    if (typeof window === "undefined") return;
    const dimensions = getWindowDimensions();
    setWindowDimensions(dimensions);
    setIsMobile(dimensions.width <= 768);
  }, []);

  useEffect(() => {
    if (typeof window === "undefined") return;

    // Set initial dimensions after mount to ensure we have the correct values
    const dimensions = getWindowDimensions();
    setWindowDimensions(dimensions);
    setIsMobile(dimensions.width <= 768);

    const throttledResize = throttle(handleResize, 100);
    window.addEventListener("resize", throttledResize, { passive: true });
    return () => {
      window.removeEventListener("resize", throttledResize);
      throttledResize.cancel();
    };
  }, [handleResize]);

  return { ...windowDimensions, isMobile };
}
