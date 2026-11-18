import { useEffect, useRef } from 'react';

export interface FilterShortcutHandlers {
  onOpenFilter: () => void;
  onToggleMine?: () => void;
  onClearAll: () => void;
}

const MODAL_OPEN = '[role="dialog"]:not([aria-modal="false"]), [role="alertdialog"]';

const isTypingContext = (target: EventTarget | null): boolean => {
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  return (
    tag === 'INPUT' ||
    tag === 'TEXTAREA' ||
    tag === 'SELECT' ||
    target.isContentEditable ||
    target.getAttribute('role') === 'textbox'
  );
};

/**
 * F opens Filter, Q toggles "Only mine", Shift+Esc clears every filter.
 * `/` is not here: the global handler focuses the page's search box when one
 * is mounted (`data-flow-filter-search`), and opens the command bar otherwise.
 *
 * Registered in the capture phase on window, like the global shortcuts. Those
 * stop a `g` from propagating, and a listener on the same node is the only
 * one that still sees it, which is how `g f` (go to Today) is told apart from
 * a lone `f` here.
 */
export function useFilterShortcuts(handlers: FilterShortcutHandlers, enabled = true): void {
  const ref = useRef(handlers);
  ref.current = handlers;

  useEffect(() => {
    if (!enabled) return undefined;
    let lastG = 0;

    const onKeyDown = (e: KeyboardEvent): void => {
      if (e.repeat || e.metaKey || e.ctrlKey || e.altKey) return;
      // A dialog or drawer on top owns the keyboard. Non-modal ones (the
      // cookie banner) do not, or the shortcuts would be dead until it is
      // dismissed.
      if (document.querySelector(MODAL_OPEN)) return;

      if (e.key === 'Escape' && e.shiftKey) {
        e.preventDefault();
        ref.current.onClearAll();
        return;
      }
      if (e.defaultPrevented || isTypingContext(e.target) || e.shiftKey) return;

      const now = Date.now();
      if (e.key === 'g') {
        lastG = now;
        return;
      }
      const afterG = now - lastG < 1200;
      lastG = 0;
      if (afterG) return;

      if (e.key === 'f') {
        e.preventDefault();
        ref.current.onOpenFilter();
      } else if (e.key === 'q' && ref.current.onToggleMine) {
        e.preventDefault();
        ref.current.onToggleMine();
      }
    };

    window.addEventListener('keydown', onKeyDown, true);
    return () => window.removeEventListener('keydown', onKeyDown, true);
  }, [enabled]);
}
