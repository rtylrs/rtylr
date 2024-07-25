import { useState, useEffect } from "react";

function getWindowDimensions() {
  const { innerWidth: width, innerHeight: height } = window;
  return {
    width,
    height,
  };
}

export default function useWindowDimensions() {
  const [windowDimensions, setWindowDimensions] = useState(
    getWindowDimensions()
  );
  const [isMobile, setIsMobile] = useState(windowDimensions.width <= 768);

  useEffect(() => {
    function handleResize() {
      const dimensions = getWindowDimensions();
      setWindowDimensions(dimensions);
      setIsMobile(dimensions.width <= 768);
    }

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return { ...windowDimensions, isMobile };
}
