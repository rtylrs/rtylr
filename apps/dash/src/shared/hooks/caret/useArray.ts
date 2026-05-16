import { useCallback, useState } from 'react';

/** Array state with stable add/remove/clear helpers. */
export function useArray<T>(initial: T[] = []): {
  items: T[];
  add: (item: T) => void;
  removeAt: (index: number) => void;
  remove: (item: T) => void;
  clear: () => void;
  set: (next: T[]) => void;
} {
  const [items, setItems] = useState<T[]>(initial);
  const add = useCallback((item: T) => setItems((prev) => [...prev, item]), []);
  const removeAt = useCallback(
    (index: number) => setItems((prev) => prev.slice(0, index).concat(prev.slice(index + 1))),
    []
  );
  const remove = useCallback((item: T) => setItems((prev) => prev.filter((x) => x !== item)), []);
  const clear = useCallback(() => setItems([]), []);
  return { items, add, removeAt, remove, clear, set: setItems };
}
