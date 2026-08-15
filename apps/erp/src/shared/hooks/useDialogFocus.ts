import { useEffect, useRef, type RefObject } from 'react';

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'area[href]',
  'button:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[contenteditable="true"]',
  '[tabindex]:not([tabindex="-1"])',
].join(',');

const getFocusableElements = (container: HTMLElement): HTMLElement[] =>
  Array.from(container.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (element) => element.getAttribute('aria-hidden') !== 'true'
  );

interface UseDialogFocusOptions {
  open: boolean;
  onClose: () => void;
  initialFocusRef?: RefObject<HTMLElement | null>;
}

/** Keeps keyboard focus inside an open transient surface and restores its trigger on close. */
export const useDialogFocus = <T extends HTMLElement = HTMLElement>({
  open,
  onClose,
  initialFocusRef,
}: UseDialogFocusOptions): RefObject<T | null> => {
  const containerRef = useRef<T>(null);
  const onCloseRef = useRef(onClose);
  const restoreFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    onCloseRef.current = onClose;
  }, [onClose]);

  useEffect(() => {
    if (!open) return;

    const container = containerRef.current;
    if (!container) return;

    restoreFocusRef.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : null;

    const initialFocus = initialFocusRef?.current;
    const focusable = getFocusableElements(container);
    (initialFocus && container.contains(initialFocus)
      ? initialFocus
      : focusable[0] || container
    ).focus();

    const handleKeyDown = (event: KeyboardEvent): void => {
      if (event.key === 'Escape') {
        event.preventDefault();
        event.stopPropagation();
        onCloseRef.current();
        return;
      }

      if (event.key !== 'Tab') return;

      const currentFocusable = getFocusableElements(container);
      if (currentFocusable.length === 0) {
        event.preventDefault();
        container.focus();
        return;
      }

      const first = currentFocusable[0];
      const last = currentFocusable[currentFocusable.length - 1];
      const activeElement = document.activeElement;

      if (event.shiftKey && (activeElement === first || !container.contains(activeElement))) {
        event.preventDefault();
        last.focus();
      } else if (
        !event.shiftKey &&
        (activeElement === last || !container.contains(activeElement))
      ) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      const restoreFocus = restoreFocusRef.current;
      restoreFocusRef.current = null;
      if (restoreFocus?.isConnected) restoreFocus.focus();
    };
  }, [initialFocusRef, open]);

  return containerRef;
};
