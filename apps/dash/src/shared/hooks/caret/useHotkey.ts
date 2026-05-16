import { useEffect } from 'react';

export interface HotkeyOptions {
  /** When true, prevent the default browser action. */
  preventDefault?: boolean;
  /** Skip when focus is in a contentEditable/input element. */
  ignoreInputs?: boolean;
}

/** Bind a global hotkey (e.g. "mod+k", "shift+/"). */
export function useHotkey(
  combo: string,
  handler: (e: KeyboardEvent) => void,
  options: HotkeyOptions = {}
): void {
  const { preventDefault = true, ignoreInputs = true } = options;

  useEffect(() => {
    const onKey = (e: KeyboardEvent): void => {
      const parts = combo.toLowerCase().split('+');
      const wantsMod = parts.includes('mod') || parts.includes('cmd') || parts.includes('ctrl');
      const wantsShift = parts.includes('shift');
      const wantsAlt = parts.includes('alt');
      const key = parts.filter((p) => !['mod', 'cmd', 'ctrl', 'shift', 'alt'].includes(p))[0];
      if (!key) return;
      const target = e.target as HTMLElement | null;
      if (
        ignoreInputs &&
        target &&
        (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)
      ) {
        return;
      }
      const mod = e.metaKey || e.ctrlKey;
      if (wantsMod !== mod) return;
      if (wantsShift !== e.shiftKey) return;
      if (wantsAlt !== e.altKey) return;
      if (e.key.toLowerCase() !== key) return;
      if (preventDefault) e.preventDefault();
      handler(e);
    };
    window.addEventListener('keydown', onKey);
    return (): void => window.removeEventListener('keydown', onKey);
  }, [combo, handler, preventDefault, ignoreInputs]);
}
