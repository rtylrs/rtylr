import React, { useState, useEffect, useRef } from 'react';
import { cn } from '../../lib/cn';
import styles from './CustomScrollbar.module.scss';

const CustomScrollbar: React.FC = () => {
  const [isVisible, setIsVisible] = useState(false);
  const [thumbHeight, setThumbHeight] = useState(0);
  const [thumbTop, setThumbTop] = useState(0);
  const [isDragging, setIsDragging] = useState(false);
  const [isClient, setIsClient] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const scrollTimeout = useRef<NodeJS.Timeout | undefined>(undefined);
  const hoverTimeout = useRef<NodeJS.Timeout | undefined>(undefined);

  useEffect(() => {
    setIsClient(true);
  }, []);

  useEffect(() => {
    if (!isClient) return;

    const updateScrollbar = () => {
      const windowHeight = window.innerHeight;
      const documentHeight = document.documentElement.scrollHeight;
      const scrollTop = window.pageYOffset;

      // Calculate thumb height as percentage of viewport (make it even shorter)
      const viewportRatio = windowHeight / documentHeight;
      const calculatedThumbHeight = Math.max(viewportRatio * windowHeight * 0.3, 30);

      // Calculate thumb position
      const maxScrollTop = documentHeight - windowHeight;
      const scrollRatio = scrollTop / maxScrollTop;
      const maxThumbTop = windowHeight - calculatedThumbHeight;
      const calculatedThumbTop = scrollRatio * maxThumbTop;

      setThumbHeight(calculatedThumbHeight);
      setThumbTop(calculatedThumbTop || 0);
    };

    const handleScroll = () => {
      setIsVisible(true);
      updateScrollbar();

      // Clear existing timeout
      if (scrollTimeout.current) {
        clearTimeout(scrollTimeout.current);
      }

      // Hide scrollbar after scroll ends
      scrollTimeout.current = setTimeout(() => {
        if (!isDragging && !isHovered) {
          setIsVisible(false);
        }
      }, 1000);
    };

    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging) return;

      e.preventDefault();

      const windowHeight = window.innerHeight;
      const documentHeight = document.documentElement.scrollHeight;
      const maxScrollTop = documentHeight - windowHeight;
      const maxThumbTop = windowHeight - thumbHeight;

      const relativeY = e.clientY - thumbHeight / 2;
      const scrollRatio = Math.max(0, Math.min(1, relativeY / maxThumbTop));
      const newScrollTop = scrollRatio * maxScrollTop;

      // Use requestAnimationFrame for smoother performance
      requestAnimationFrame(() => {
        window.scrollTo(0, newScrollTop);
      });
    };

    const handleMouseUp = () => {
      setIsDragging(false);
    };

    const handleGlobalMouseMove = (e: MouseEvent) => {
      const isNearRightEdge = window.innerWidth - e.clientX <= 30;

      if (isNearRightEdge && !isHovered) {
        setIsHovered(true);
        setIsVisible(true);
      } else if (!isNearRightEdge && isHovered && !isDragging) {
        if (hoverTimeout.current) {
          clearTimeout(hoverTimeout.current);
        }
        hoverTimeout.current = setTimeout(() => {
          setIsHovered(false);
          if (!isDragging) {
            setIsVisible(false);
          }
        }, 500);
      }
    };

    window.addEventListener('scroll', handleScroll);
    window.addEventListener('resize', updateScrollbar);
    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mousemove', handleGlobalMouseMove);
    window.addEventListener('mouseup', handleMouseUp);

    // Initial calculation
    updateScrollbar();

    return () => {
      window.removeEventListener('scroll', handleScroll);
      window.removeEventListener('resize', updateScrollbar);
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mousemove', handleGlobalMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
      if (scrollTimeout.current) {
        clearTimeout(scrollTimeout.current);
      }
      if (hoverTimeout.current) {
        clearTimeout(hoverTimeout.current);
      }
    };
  }, [isClient, isDragging, thumbHeight, isHovered]);

  const handleThumbMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    setIsDragging(true);
    setIsVisible(true);
  };

  const handleTrackClick = (e: React.MouseEvent) => {
    // Prevent click when dragging thumb
    if (e.target !== e.currentTarget) return;

    const rect = e.currentTarget.getBoundingClientRect();
    const clickY = e.clientY - rect.top;
    const windowHeight = window.innerHeight;
    const documentHeight = document.documentElement.scrollHeight;
    const maxScrollTop = documentHeight - windowHeight;
    const maxThumbTop = windowHeight - thumbHeight;

    const scrollRatio = Math.max(0, Math.min(1, (clickY - thumbHeight / 2) / maxThumbTop));
    const newScrollTop = scrollRatio * maxScrollTop;

    window.scrollTo({
      top: newScrollTop,
      behavior: 'smooth',
    });
  };

  // Don't render on server or if content doesn't overflow
  if (!isClient) {
    return null;
  }

  if (document.documentElement.scrollHeight <= window.innerHeight) {
    return null;
  }

  return (
    <div
      className={cn(styles.scrollbar, isVisible && styles.visible, isHovered && styles.hovered)}
      onClick={handleTrackClick}
    >
      <div
        className={styles.thumb}
        style={{
          height: `${thumbHeight}px`,
          top: `${thumbTop}px`,
        }}
        onMouseDown={handleThumbMouseDown}
      />
    </div>
  );
};

export default CustomScrollbar;
