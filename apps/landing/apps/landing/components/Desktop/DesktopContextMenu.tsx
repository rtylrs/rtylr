import React, { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useDesktopStore } from '../../stores/useDesktopStore';

interface DesktopContextMenuProps {
  x: number;
  y: number;
  onClose: () => void;
}

export const DesktopContextMenu: React.FC<DesktopContextMenuProps> = ({ x, y, onClose }) => {
  const addWindow = useDesktopStore((s) => s.addWindow);
  const closeAllWindows = useDesktopStore((s) => s.closeAllWindows);
  const menuRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ x, y });

  useLayoutEffect(() => {
    const el = menuRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    setPos({
      x: Math.max(8, Math.min(x, window.innerWidth - rect.width - 8)),
      y: Math.max(8, Math.min(y, window.innerHeight - rect.height - 8)),
    });
  }, [x, y]);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        onClose();
      }
    };
    const keyHandler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('mousedown', handler);
    window.addEventListener('keydown', keyHandler);
    return () => {
      window.removeEventListener('mousedown', handler);
      window.removeEventListener('keydown', keyHandler);
    };
  }, [onClose]);

  const item = (label: string, action: () => void) => (
    <div
      role="menuitem"
      tabIndex={0}
      onClick={() => {
        action();
        onClose();
      }}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          action();
          onClose();
        }
      }}
      style={{
        padding: '7px 14px',
        cursor: 'pointer',
        fontSize: 13,
        color: 'rgba(255,255,255,0.85)',
        borderRadius: 5,
        whiteSpace: 'nowrap',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLDivElement).style.background = 'rgba(255,255,255,0.08)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLDivElement).style.background = 'transparent';
      }}
    >
      {label}
    </div>
  );

  return (
    <div
      ref={menuRef}
      role="menu"
      style={{
        position: 'fixed',
        top: pos.y,
        left: pos.x,
        zIndex: 8900,
        background: 'rgba(18,18,22,0.96)',
        backdropFilter: 'blur(16px)',
        border: '1px solid rgba(255,255,255,0.1)',
        borderRadius: 8,
        padding: 4,
        minWidth: 188,
        boxShadow: '0 8px 32px rgba(0,0,0,0.6)',
      }}
    >
      {item('New Window', () => addWindow('landing'))}
      {item('Change Wallpaper', () => addWindow('settings'))}
      {item('About This System', () => addWindow('about'))}
      <div
        style={{
          height: 1,
          background: 'rgba(255,255,255,0.08)',
          margin: '4px 0',
        }}
      />
      {item('Close All Windows', () => closeAllWindows())}
    </div>
  );
};
