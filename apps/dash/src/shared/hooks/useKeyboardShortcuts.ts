import { useEffect, useRef } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useCommandPalette } from '../stores/commandPaletteStore';

/**
 * Registers global keyboard shortcuts:
 *  - cmd+k / ctrl+k : toggle command palette
 *  - g t            : go to /team
 *  - g a            : go to /apps
 *  - g s            : go to /settings
 *  - g d            : go to /
 *  - g n            : go to /notifications
 *  - n              : new (contextual, here we push /team/invite as a sensible default)
 *  - ?              : open shortcuts help (toggles palette with "?")
 */
export const useKeyboardShortcuts = () => {
  const navigate = useNavigate();
  const { toggle, setOpen } = useCommandPalette();
  const chordRef = useRef<{ key: string; at: number } | null>(null);

  useEffect(() => {
    const isEditable = (el: Element | null): boolean => {
      if (!el) return false;
      const tag = el.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return true;
      return (el as HTMLElement).isContentEditable;
    };

    const handler = (e: KeyboardEvent) => {
      // cmd/ctrl+k
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        toggle();
        return;
      }

      if (isEditable(document.activeElement)) return;

      // "?" opens the palette as a help shortcut
      if (e.key === '?') {
        e.preventDefault();
        setOpen(true);
        return;
      }

      // "n" = new (invite)
      if (e.key.toLowerCase() === 'n' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        e.preventDefault();
        navigate({ to: '/team/invite' });
        return;
      }

      // chord: g + [letter]
      if (e.key.toLowerCase() === 'g' && !e.metaKey && !e.ctrlKey && !e.altKey) {
        chordRef.current = { key: 'g', at: Date.now() };
        return;
      }

      if (chordRef.current && chordRef.current.key === 'g') {
        const elapsed = Date.now() - chordRef.current.at;
        if (elapsed > 1500) {
          chordRef.current = null;
          return;
        }
        const k = e.key.toLowerCase();
        const routes: Record<string, string> = {
          d: '/',
          t: '/team',
          a: '/apps',
          s: '/settings',
          n: '/notifications',
          b: '/billing',
          i: '/integrations',
          p: '/permissions',
          k: '/api-keys',
          h: '/audit',
          u: '/audit',
          r: '/roles',
        };
        if (routes[k]) {
          e.preventDefault();
          navigate({ to: routes[k] as string });
        }
        chordRef.current = null;
      }
    };

    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [navigate, toggle, setOpen]);
};
