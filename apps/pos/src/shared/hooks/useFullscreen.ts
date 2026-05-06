import { useEffect, useCallback } from 'react';
import { getCurrentWindow } from '@tauri-apps/api/window';
import { useFullscreenStore } from '../stores/fullscreenStore';
import { usePlatform } from './usePlatform';

/**
 * Hook to manage fullscreen state with Tauri window API
 * Only works in Tauri environment, gracefully handles web environment
 */
export const useFullscreen = () => {
  const { isTauri } = usePlatform();
  const { isFullscreen, setFullscreen, toggleFullscreen } = useFullscreenStore();

  // Apply fullscreen state to the window
  const applyFullscreen = useCallback(
    async (shouldBeFullscreen: boolean) => {
      if (!isTauri) {
        return;
      }

      try {
        await getCurrentWindow().setFullscreen(shouldBeFullscreen);
        setFullscreen(shouldBeFullscreen);
      } catch (error) {}
    },
    [isTauri, setFullscreen]
  );

  // Toggle fullscreen state
  const toggle = useCallback(async () => {
    const newState = !isFullscreen;
    await applyFullscreen(newState);
    toggleFullscreen();
  }, [isFullscreen, applyFullscreen, toggleFullscreen]);

  // Apply stored fullscreen state on mount
  useEffect(() => {
    if (isTauri) {
      applyFullscreen(isFullscreen);
    }
  }, [isTauri]); // Only run on mount when platform is detected

  return {
    isFullscreen,
    toggle,
    setFullscreen: applyFullscreen,
  };
};
