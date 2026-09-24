import React, { createContext, useContext, useEffect, useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { CommandPalette } from './CommandPalette';

interface ShortcutsContextValue {
  openPalette: () => void;
}

const ShortcutsContext = createContext<ShortcutsContextValue>({ openPalette: () => {} });

export function useShortcuts() {
  return useContext(ShortcutsContext);
}

export const ShortcutsProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [paletteOpen, setPaletteOpen] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const isTyping = () => {
      const el = document.activeElement as HTMLElement | null;
      // A focused <select> is mid-choice: a letter shortcut would navigate
      // away from the open dropdown, so treat it like a text field (#943).
      return (
        el &&
        (el.tagName === 'INPUT' ||
          el.tagName === 'TEXTAREA' ||
          el.tagName === 'SELECT' ||
          el.isContentEditable)
      );
    };
    const handler = (e: KeyboardEvent) => {
      // cmd+k / ctrl+k opens palette (always works)
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        // Caret also owns Cmd/Ctrl+K. Handle the shortcut in the capture
        // phase so this palette is the single owner for CRM.
        e.stopImmediatePropagation();
        setPaletteOpen((o) => !o);
        return;
      }
      if (isTyping()) return;
      if (e.key === '/') {
        e.preventDefault();
        setPaletteOpen(true);
        return;
      }
      if (e.key === 'n' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        navigate({ to: '/customers/new' });
        return;
      }
      if (e.shiftKey && e.key.toLowerCase() === 's') {
        e.preventDefault();
        navigate({ to: '/segments' });
        return;
      }
      if (e.shiftKey && e.key.toLowerCase() === 'c') {
        e.preventDefault();
        navigate({ to: '/campaigns' });
        return;
      }
    };
    window.addEventListener('keydown', handler, true);
    return () => window.removeEventListener('keydown', handler, true);
  }, [navigate]);

  return (
    <ShortcutsContext.Provider value={{ openPalette: () => setPaletteOpen(true) }}>
      {children}
      <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
    </ShortcutsContext.Provider>
  );
};
