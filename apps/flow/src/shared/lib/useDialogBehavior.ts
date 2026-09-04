import { useCallback, useEffect, useRef } from 'react';

/**
 * Esc, focus and scroll for a modal dialog, in one place.
 *
 * Flow's dialogs each put `onKeyDown` on the overlay element:
 *
 * ```tsx
 * <div role="dialog" aria-modal="true" onKeyDown={(e) => { if (e.key === 'Escape') onClose(); }}>
 * ```
 *
 * A `div` with no `tabIndex` cannot hold focus, and these dialogs never moved
 * focus into themselves on open, so the keydown landed on `document.body`,
 * never bubbled through the overlay, and Esc did nothing. The two defects were
 * one: no focus inside meant no key events, which meant no Esc. Listening on
 * the document instead makes the handler independent of where focus happens to
 * be.
 *
 * Tab is cycled within the panel rather than merely started there, because
 * `aria-modal="true"` tells assistive technology the rest of the page is
 * inert while the browser still lets Tab walk out of it.
 */

const FOCUSABLE = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

function focusableWithin(panel: HTMLElement): HTMLElement[] {
  return [...panel.querySelectorAll<HTMLElement>(FOCUSABLE)].filter(
    (element) => element.offsetParent !== null || element === document.activeElement
  );
}

export interface DialogBehavior<T extends HTMLElement> {
  /** Put this on the panel, not the overlay: it is what focus moves into. */
  panelRef: React.RefObject<T | null>;
}

export function useDialogBehavior<T extends HTMLElement>(
  open: boolean,
  onClose: () => void
): DialogBehavior<T> {
  const panelRef = useRef<T | null>(null);
  const restoreTo = useRef<HTMLElement | null>(null);

  // Held in a ref so the document listener does not need re-attaching each
  // time a parent re-renders with a new callback identity.
  const closeRef = useRef(onClose);
  useEffect(() => {
    closeRef.current = onClose;
  }, [onClose]);

  const trapTab = useCallback((event: KeyboardEvent) => {
    const panel = panelRef.current;
    if (!panel) return;
    const focusable = focusableWithin(panel);
    if (focusable.length === 0) {
      // Nothing to land on, so keep Tab inside rather than sending the user
      // out to the page behind an aria-modal dialog.
      event.preventDefault();
      panel.focus();
      return;
    }
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const active = document.activeElement;
    if (event.shiftKey && (active === first || active === panel)) {
      event.preventDefault();
      last.focus();
      return;
    }
    if (!event.shiftKey && active === last) {
      event.preventDefault();
      first.focus();
    }
  }, []);

  useEffect(() => {
    if (!open) return;

    restoreTo.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;

    const panel = panelRef.current;
    if (panel) {
      const focusable = focusableWithin(panel);
      // The panel itself when it holds nothing focusable, so Esc and the trap
      // still have somewhere to work from.
      (focusable[0] ?? panel).focus();
    }

    const onKeyDown = (event: KeyboardEvent): void => {
      if (event.key === 'Escape') {
        event.stopPropagation();
        closeRef.current();
        return;
      }
      if (event.key === 'Tab') trapTab(event);
    };

    document.addEventListener('keydown', onKeyDown, true);
    return () => {
      document.removeEventListener('keydown', onKeyDown, true);
      // Only if focus is still inside the dialog: a close that already moved
      // focus somewhere deliberate should not be overridden.
      //
      // `panel` is the element captured when the effect ran, not a fresh read
      // of the ref. React can clear the ref before this cleanup, and on
      // unmount the captured node is detached, so `contains` returns false and
      // the `document.body` branch below is what restores focus.
      const active = document.activeElement;
      const stillInside = panel != null && active instanceof Node && panel.contains(active);
      if ((stillInside || active === document.body) && restoreTo.current?.isConnected) {
        restoreTo.current.focus();
      }
    };
  }, [open, trapTab]);

  return { panelRef };
}
