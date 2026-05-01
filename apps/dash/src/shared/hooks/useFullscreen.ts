import { useEffect, useCallback } from 'react';
import { useFullscreenStore } from '../stores/fullscreenStore';

/**
 * Hook to manage fullscreen state for web environment
 * Note: Actual fullscreen functionality removed (was Tauri-specific)
 */
export const useFullscreen = () => {
  const { isFullscreen, setFullscreen, toggleFullscreen } = useFullscreenStore();

  // Apply fullscreen state to the window
  const applyFullscreen = useCallback(
    async (shouldBeFullscreen: boolean) => {
      // Just update the store state for UI consistency
      setFullscreen(shouldBeFullscreen);
    },
    [setFullscreen]
  );

  // Toggle fullscreen state
  const toggle = useCallback(async () => {
    const newState = !isFullscreen;
    await applyFullscreen(newState);
    toggleFullscreen();
  }, [isFullscreen, applyFullscreen, toggleFullscreen]);

  // Apply stored fullscreen state on mount
  useEffect(() => {
    // No-op in web-only version
  }, []); // Only run on mount

  return {
    isFullscreen,
    toggle,
    setFullscreen: applyFullscreen,
  };
};
