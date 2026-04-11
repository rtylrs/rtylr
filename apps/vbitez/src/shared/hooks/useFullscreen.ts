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
        console.warn('Fullscreen control is only available in Tauri desktop app');
        return;
      }

      try {
        await getCurrentWindow().setFullscreen(shouldBeFullscreen);
        setFullscreen(shouldBeFullscreen);
      } catch (error) {
        console.error('Failed to set fullscreen:', error);
      }
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
