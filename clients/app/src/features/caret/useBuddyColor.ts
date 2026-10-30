import { useCallback, useEffect, useState } from 'react';

export const BUDDY_COLORS = {
  ink: '',
  tangerine: '#ff6a3d',
  sky: '#3b82f6',
  mint: '#10b981',
  grape: '#8b5cf6',
  rose: '#f43f5e',
} as const;
export type BuddyColor = keyof typeof BUDDY_COLORS;

const EVENT = 'caret-buddy-color';

const read = (key: string): BuddyColor => {
  const value = localStorage.getItem(key);
  return value && value in BUDDY_COLORS ? (value as BuddyColor) : 'ink';
};

/** The person's chosen Buddy colour, shared by every sprite on the page. */
export function useBuddyColor(scope: string): [BuddyColor, (color: BuddyColor) => void] {
  const key = `caret-buddy-color:${scope}`;
  const [color, setColor] = useState<BuddyColor>(() => read(key));
  useEffect(() => {
    const sync = () => setColor(read(key));
    sync();
    window.addEventListener(EVENT, sync);
    window.addEventListener('storage', sync);
    return () => {
      window.removeEventListener(EVENT, sync);
      window.removeEventListener('storage', sync);
    };
  }, [key]);
  const choose = useCallback(
    (next: BuddyColor) => {
      localStorage.setItem(key, next);
      window.dispatchEvent(new Event(EVENT));
    },
    [key]
  );
  return [color, choose];
}
