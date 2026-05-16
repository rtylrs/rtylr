import { useMediaQuery } from '@voxire/caret';

/** True when the user's OS prefers dark mode. */
export function useDarkMode(): boolean {
  return useMediaQuery('(prefers-color-scheme: dark)');
}
