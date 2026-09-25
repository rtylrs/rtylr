import { useCallback, useState } from 'react';

/** Wrap navigator.clipboard with a 1.5s "copied" indicator. */
export function useClipboard(): {
  copy: (text: string) => Promise<boolean>;
  copied: boolean;
} {
  const [copied, setCopied] = useState(false);

  const copy = useCallback(async (text: string): Promise<boolean> => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
      return true;
    } catch {
      return false;
    }
  }, []);

  return { copy, copied };
}
