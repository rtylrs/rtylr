import { useCallback, useState } from 'react';
import type { Task } from '../stores/flowStore';

export const isClosedTask = (task: Task): boolean =>
  task.status === 'done' || task.status === 'cancelled';

/**
 * "Show completed" for a cross-project list, remembered per list. Done and
 * cancelled work is left out by default because these lists are about what is
 * still open.
 */
export function useShowCompleted(key: string): [boolean, (value: boolean) => void] {
  const storageKey = `flow:${key}:show-completed`;
  const [value, setValue] = useState(() => {
    try {
      return localStorage.getItem(storageKey) === '1';
    } catch {
      return false;
    }
  });
  const set = useCallback(
    (next: boolean) => {
      setValue(next);
      try {
        localStorage.setItem(storageKey, next ? '1' : '0');
      } catch {
        // A preference that does not stick is not worth an error.
      }
    },
    [storageKey]
  );
  return [value, set];
}

/**
 * Whether closed tasks belong in the list: the toggle is on, or the status
 * filter asks for a closed status (which would otherwise show nothing).
 */
export function includesClosed(showCompleted: boolean, statuses: readonly string[]): boolean {
  return showCompleted || statuses.some((s) => s === 'done' || s === 'cancelled');
}
