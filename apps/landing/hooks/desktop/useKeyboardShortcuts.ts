import { useEffect } from 'react';
import { useDesktopStore } from '../../stores/useDesktopStore';

export function useKeyboardShortcuts() {
  const closeAllWindows = useDesktopStore((s) => s.closeAllWindows);
  const setExperience = useDesktopStore((s) => s.setExperience);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.shiftKey && e.key === 'X') {
        e.preventDefault();
        closeAllWindows();
      }
      if (e.shiftKey && e.key === 'M') {
        e.preventDefault();
        setExperience('website');
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [closeAllWindows, setExperience]);
}
