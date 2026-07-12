import { useCallback } from 'react';
import { useFullscreenStore } from '../stores/fullscreenStore';

/**
 * Hook to manage fullscreen state for web environment
 * Note: Actual fullscreen functionality removed (was Tauri-specific)
 */
export const useFullscreen = () => {
  const { isFullscreen, setFullscreen } = useFullscreenStore();

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
    await applyFullscreen(!isFullscreen);
  }, [isFullscreen, applyFullscreen]);

  return {
    isFullscreen,
    toggle,
    setFullscreen: applyFullscreen,
  };
};
