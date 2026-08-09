import { useEffect, useLayoutEffect, useRef } from 'react';

interface SavedScrollPosition {
  scrollTop: number;
  pinnedToBottom: boolean;
}

const scrollPositions = new Map<string, SavedScrollPosition>();
const DEFAULT_SCROLL_KEY = 'caret-default-scroll';

function readScrollPosition(element: HTMLElement): SavedScrollPosition {
  const distance = element.scrollHeight - element.scrollTop - element.clientHeight;
  return {
    scrollTop: element.scrollTop,
    pinnedToBottom: distance < 96,
  };
}

/**
 * Returns a ref to attach to the scroll container, and an effect that keeps
 * the container pinned to the bottom while the user hasn't scrolled away.
 * Triggered on changes to `deps`.
 */
export function useScrollToBottom<T extends HTMLElement>(
  deps: unknown[],
  scrollKey = DEFAULT_SCROLL_KEY
): React.RefObject<T | null> {
  const ref = useRef<T>(null);
  const activeKeyRef = useRef(scrollKey);
  const previousKeyRef = useRef(scrollKey);
  const pinnedToBottomRef = useRef(true);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    const savePosition = (): void => {
      const position = readScrollPosition(el);
      pinnedToBottomRef.current = position.pinnedToBottom;
      scrollPositions.set(activeKeyRef.current, position);
    };

    savePosition();
    el.addEventListener('scroll', savePosition, { passive: true });
    return (): void => el.removeEventListener('scroll', savePosition);
  }, []);

  useLayoutEffect(() => {
    const el = ref.current;
    if (!el) return;

    activeKeyRef.current = scrollKey;
    const keyChanged = previousKeyRef.current !== scrollKey;
    const saved = scrollPositions.get(scrollKey);

    if (keyChanged) {
      if (!saved || saved.pinnedToBottom) {
        el.scrollTop = el.scrollHeight;
        pinnedToBottomRef.current = true;
      } else {
        const maxScrollTop = Math.max(0, el.scrollHeight - el.clientHeight);
        el.scrollTop = Math.min(saved.scrollTop, maxScrollTop);
        pinnedToBottomRef.current = false;
      }
      previousKeyRef.current = scrollKey;
    } else if (saved?.pinnedToBottom ?? pinnedToBottomRef.current) {
      el.scrollTop = el.scrollHeight;
      pinnedToBottomRef.current = true;
    }

    scrollPositions.set(scrollKey, readScrollPosition(el));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scrollKey, ...deps]);

  return ref;
}
