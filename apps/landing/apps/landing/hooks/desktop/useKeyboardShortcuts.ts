import { useEffect } from 'react';
import { useDesktopStore } from '../../stores/useDesktopStore';

const isEditableTarget = (target: EventTarget | null): boolean =>
  target instanceof HTMLElement &&
  (target.tagName === 'INPUT' ||
    target.tagName === 'TEXTAREA' ||
    target.tagName === 'SELECT' ||
    target.isContentEditable);

export function useKeyboardShortcuts() {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const store = useDesktopStore.getState();
      const mod = e.metaKey || e.ctrlKey;

      if (mod && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        store.setSpotlightOpen(!store.spotlightOpen);
        return;
      }
      if (mod && e.key === '/') {
        e.preventDefault();
        store.addWindow('shortcuts');
        return;
      }

      if (mod || e.altKey || isEditableTarget(e.target)) return;

      if (e.shiftKey && e.key === 'X') {
        e.preventDefault();
        store.closeAllWindows();
        return;
      }
      if (e.shiftKey && e.key === 'M') {
        e.preventDefault();
        store.setExperience('website');
        return;
      }
      if (e.shiftKey && e.key === 'W') {
        const focused = store.windows
          .filter((w) => !w.minimized)
          .sort((a, b) => b.zIndex - a.zIndex)[0];
        if (focused) {
          e.preventDefault();
          store.closeWindow(focused.id);
        }
        return;
      }
      if (e.shiftKey && e.key === 'H') {
        const focused = store.windows
          .filter((w) => !w.minimized)
          .sort((a, b) => b.zIndex - a.zIndex)[0];
        if (focused) {
          e.preventDefault();
          store.minimizeWindow(focused.id);
        }
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);
}
