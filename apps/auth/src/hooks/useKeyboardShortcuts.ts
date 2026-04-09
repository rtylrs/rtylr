import { useEffect } from 'react';

export interface KeyboardShortcutHandlers {
  onEnter?: () => void;
  onEscape?: () => void;
  onSubmit?: () => void;
}

/**
 * Binds global Enter / Escape handlers. Enter triggers onSubmit or onEnter,
 * Escape triggers onEscape. Automatically ignores Enter when focus is on a
 * textarea or button so it doesn't hijack normal form flow.
 */
export const useKeyboardShortcuts = (handlers: KeyboardShortcutHandlers) => {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      const tag = target?.tagName.toLowerCase();
      const isFormControl =
        tag === 'input' || tag === 'select' || tag === 'textarea';

      if (e.key === 'Enter' && (handlers.onSubmit || handlers.onEnter)) {
        if (tag === 'textarea') return;
        if (isFormControl && handlers.onSubmit) {
          // Forms auto-submit on Enter; don't double-trigger.
          return;
        }
        handlers.onEnter?.();
      }
      if (e.key === 'Escape' && handlers.onEscape) {
        handlers.onEscape();
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [handlers]);
};
