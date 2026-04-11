import React, { createContext, useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
      return el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable);
    };
    const handler = (e: KeyboardEvent) => {
      // cmd+k / ctrl+k opens palette (always works)
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
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
        navigate('/customers/new');
        return;
      }
      if (e.shiftKey && e.key.toLowerCase() === 's') {
        e.preventDefault();
        navigate('/segments');
        return;
      }
      if (e.shiftKey && e.key.toLowerCase() === 'c') {
        e.preventDefault();
        navigate('/campaigns');
        return;
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [navigate]);

  return (
    <ShortcutsContext.Provider value={{ openPalette: () => setPaletteOpen(true) }}>
      {children}
      <CommandPalette open={paletteOpen} onClose={() => setPaletteOpen(false)} />
    </ShortcutsContext.Provider>
  );
};
