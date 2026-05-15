import { useCallback, useEffect, useRef } from 'react';
import { useCaretSettingsStore } from '../stores/settingsStore';

/**
 * Drag the left edge of the Caret panel to resize. Width is persisted via the
 * settings store. The drag handle is whatever element receives the returned
 * `onMouseDown`.
 */
export function usePanelResize(): {
  onMouseDown: (e: React.MouseEvent) => void;
  width: number;
} {
  const width = useCaretSettingsStore((s) => s.panelWidth);
  const setPanelWidth = useCaretSettingsStore((s) => s.setPanelWidth);
  const draggingRef = useRef(false);

  const onMouseMove = useCallback(
    (e: MouseEvent): void => {
      if (!draggingRef.current) return;
      // Drag from the left edge of a right-anchored panel: width grows as
      // mouse moves left (smaller clientX), shrinks as it moves right.
      const next = window.innerWidth - e.clientX;
      setPanelWidth(next);
    },
    [setPanelWidth]
  );

  const onMouseUp = useCallback((): void => {
    draggingRef.current = false;
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
  }, []);

  useEffect(() => {
    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
    return (): void => {
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
    };
  }, [onMouseMove, onMouseUp]);

  const onMouseDown = useCallback((e: React.MouseEvent): void => {
    e.preventDefault();
    draggingRef.current = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
  }, []);

  return { onMouseDown, width };
}
